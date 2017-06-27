//go:generate bash ./g_version.sh
package main

import (
	"bytes"
	"fmt"
	"log/syslog"
	"os"
	"os/exec"
	"path"

	"log"

	"gopkg.in/alecthomas/kingpin.v1"
)

var (
	appName   = path.Base(os.Args[0])
	app       = kingpin.New(appName, "A command-line checker for Disk Health checks using smartctl, by CrossEngage")
	checkName = app.Flag("name", "check name").Default(appName).String()
	smartCtl  = app.Flag("smartctl", "Path of smartctl").Default("/usr/sbin/smartctl").String()
)

func main() {
	app.Version(version)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	_, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	slog, err := syslog.New(syslog.LOG_NOTICE|syslog.LOG_DAEMON, appName)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(slog)

	cmd := exec.Command(*smartCtl, "--output-event-state", "--comma-separated-output", "--no-header-output")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed running `%s` with error `%s`\n", *smartCtl, err)
	}

	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	slog.Debug(fmt.Sprintf("%s: stdout `%s`, stderr `%s`", *smartCtl, outStr, errStr))

	// outStr = strings.TrimSpace(outStr)
	// lines := strings.Split(outStr, "\n")
	// if len(lines) >= 0 {
	// 	for _, line := range lines {
	// 		ev, err := ,,,(line)
	// 		if err != nil {
	// 			slog.Err(fmt.Sprintf("Could not parse `%s` stdout: `%s`", line, outStr))
	// 		}
	// 		fmt.Println(ev.InfluxDB(*checkName, hostname))
	// 	}
	// } else {
	// 	ev := ...()
	// 	fmt.Println(ev.InfluxDB(*checkName, hostname))
	// }
}
