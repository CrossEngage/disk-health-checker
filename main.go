//go:generate bash ./g_version.sh
package main

import (
	"bytes"
	"fmt"
	"log"
	"log/syslog"
	"os"
	"os/exec"
	"path"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	appName   = path.Base(os.Args[0])
	app       = kingpin.New(appName, "A command-line checker for Disk Health checks using smartctl, by CrossEngage")
	checkName = app.Flag("name", "check name").Default(appName).String()
	debug     = app.Flag("debug", "if set, enables debug logs").Default("false").Bool()
	stderr    = app.Flag("stderr", "if set, enables logging to stderr instead of syslog").Default("false").Bool()
	smartCtl  = app.Flag("smartctl", "Path of smartctl").Default("/usr/sbin/smartctl").String()

	// https://en.wikipedia.org/wiki/S.M.A.R.T.#Known_ATA_S.M.A.R.T._attributes
	attrIDs = app.Flag("attrs", "SMART Attribute IDs to return").Default(
		"1", "2", "3", "5", "7", "8", "9", "10", "12", "171", "172", "173",
		"174", "190", "194", "197", "198", "199", "231", "233").Ints()
)

func main() {
	app.Version(version)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *stderr {
		log.SetOutput(os.Stderr)
	} else {
		slog, err := syslog.New(syslog.LOG_NOTICE|syslog.LOG_DAEMON, appName)
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(slog)

	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	stdOut, _, err := smartctl(*debug, "--scan")
	if err != nil {
		log.Fatal(err)
	}

	devices := parseSMARTCtlScan(stdOut)
	for _, device := range devices {
		stdOut, _, err := smartctl(*debug, "-i", "-H", device.Path, "-d", device.Type)
		if err != nil {
			log.Println(err)
		}

		info := parseSMARTCtlInfo(stdOut)
		fmt.Printf("%s,host=%s,disk=%s,type=%s ", *checkName, hostname, device.Path, strings.Replace(device.Type, ",", "_", -1))
		values := []string{fmt.Sprintf(`disk_status="%s"`, info.Health)}

		if info.SMARTSupport {
			stdOut, _, err = smartctl(*debug, "-A", device.Path, "-d", device.Type)
			if err != nil {
				log.Println(err)
			}
			attrs := parseAttributeList(stdOut)
			for _, attr := range attrs {
				// TODO replace this by something more efficient
				for _, id := range *attrIDs {
					if attr.ID == id {
						values = append(values, attr.String(true, false))
						continue
					}
				}
			}
		}

		fmt.Println(strings.Join(values, ","))
	}
}

func smartctl(debug bool, args ...string) (string, string, error) {
	cmd := exec.Command(*smartCtl, args...)
	if debug {
		log.Printf("Running `%s with args: %v", *smartCtl, args)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	if debug {
		log.Printf("%s: stdout `%s`, stderr `%s`", *smartCtl, strings.TrimSpace(outStr), strings.TrimSpace(errStr))
	}
	return outStr, errStr, err
}
