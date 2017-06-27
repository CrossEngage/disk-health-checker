package main

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type deviceInfo struct {
	Raw  string
	Path string
	Type string
}

var deviceInfoRgx = regexp.MustCompile(`(\S+) \-d (\S+)`)

func parseSMARTCtlScan(out string) []deviceInfo {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	devices := make([]deviceInfo, 0, len(lines))
	for _, line := range lines {
		tmp := strings.Split(line, "#")
		di := deviceInfo{Raw: tmp[0]}
		if m := deviceInfoRgx.FindAllStringSubmatch(di.Raw, -1); m != nil {
			di.Path = m[0][1]
			di.Type = m[0][2]
			if di.Type == "scsi" {
				di.Type = "auto"
			}
			devices = append(devices, di)
		}
	}
	return devices
}

type smartCtlInfo struct {
	DeviceModel           string
	SerialNumber          string
	LUWWNDeviceID         string
	FirmwareVersion       string
	UserCapacity          string
	UserCapacityBytes     int64
	SectorSizes           string
	LogicalSectorSizes    int
	PhysicalSectorSizes   int
	RotationRate          string
	RotationRateRPM       int
	IsSSD                 bool
	ATAVersion            string
	SATAVersion           string
	SMARTSupportIs        string
	SMARTSupport          bool
	Vendor                string
	Product               string
	Revision              string
	LogicalBlockSize      string
	LogicalBlockSizeBytes int
	LogicalUnitID         string
	DeviceType            string
	Health                string
	Healthy               bool
}

var (
	separateSectorSizesRgx     = regexp.MustCompile(`^(\d+) bytes logical, (\d+) bytes physical$`)
	sameSectorSizesRgx         = regexp.MustCompile(`^(\d+) bytes logical\/physical$`)
	userCapacityRgx            = regexp.MustCompile(`^([,\d]+) bytes \[.+?\]$`)
	rpmRgx                     = regexp.MustCompile(`^([\d]+) rpm$`)
	bytesRgx                   = regexp.MustCompile(`^(\d+) bytes$`)
	columnsRgx                 = regexp.MustCompile(`\s+`)
	influxDBFieldNameFilterRgx = regexp.MustCompile(`[^A-Za-z0-9]`)
)

func parseSMARTCtlInfo(out string) *smartCtlInfo {
	info := &smartCtlInfo{Health: "UNSUPPORTED"}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		sliced := strings.SplitN(line, ":", 2)
		switch sliced[0] {
		case "Device Model":
			info.DeviceModel = strings.TrimSpace(sliced[1])
		case "Serial Number":
			info.SerialNumber = strings.TrimSpace(sliced[1])
		case "LU WWN Device Id":
			info.LUWWNDeviceID = strings.TrimSpace(sliced[1])
		case "Firmware Version":
			info.FirmwareVersion = strings.TrimSpace(sliced[1])
		case "User Capacity":
			info.UserCapacity = strings.TrimSpace(sliced[1])
			if m := userCapacityRgx.FindAllStringSubmatch(info.UserCapacity, -1); m != nil {
				info.UserCapacityBytes, _ = strconv.ParseInt(strings.Replace(m[0][1], ",", "", -1), 10, 64)
			}
		case "Sector Size":
			info.SectorSizes = strings.TrimSpace(sliced[1])
			if m := sameSectorSizesRgx.FindAllStringSubmatch(info.SectorSizes, -1); m != nil {
				info.LogicalSectorSizes, _ = strconv.Atoi(m[0][1])
				info.PhysicalSectorSizes = info.LogicalSectorSizes
			}
		case "Sector Sizes":
			info.SectorSizes = strings.TrimSpace(sliced[1])
			if m := separateSectorSizesRgx.FindAllStringSubmatch(info.SectorSizes, -1); m != nil {
				info.LogicalSectorSizes, _ = strconv.Atoi(m[0][1])
				info.PhysicalSectorSizes, _ = strconv.Atoi(m[0][2])
			}
		case "Rotation Rate":
			info.RotationRate = strings.TrimSpace(sliced[1])
			if m := rpmRgx.FindAllStringSubmatch(info.RotationRate, -1); m != nil {
				info.RotationRateRPM, _ = strconv.Atoi(m[0][1])
			}
			info.IsSSD = info.RotationRate == "Solid State Device"
		case "ATA Version is":
			info.ATAVersion = strings.TrimSpace(sliced[1])
		case "SATA Version is":
			info.SATAVersion = strings.TrimSpace(sliced[1])
		case "SMART support is":
			info.SMARTSupportIs = strings.TrimSpace(sliced[1])
			info.SMARTSupport = info.SMARTSupportIs == "Enabled"
		case "Vendor":
			info.Vendor = strings.TrimSpace(sliced[1])
		case "Product":
			info.Product = strings.TrimSpace(sliced[1])
		case "Revision":
			info.Revision = strings.TrimSpace(sliced[1])
		case "Logical block size":
			info.LogicalBlockSize = strings.TrimSpace(sliced[1])
			if m := bytesRgx.FindAllStringSubmatch(info.LogicalBlockSize, -1); m != nil {
				info.LogicalBlockSizeBytes, _ = strconv.Atoi(m[0][1])
			}
		case "Logical Unit id":
			info.LogicalUnitID = strings.TrimSpace(sliced[1])
		case "Device type":
			info.DeviceType = strings.TrimSpace(sliced[1])
		case "SMART overall-health self-assessment test result":
			info.Health = strings.TrimSpace(sliced[1])
			info.Healthy = info.Health == "PASSED"
		}
	}

	return info
}

type smartAttribute struct {
	ID            int    ``
	Name          string ``
	Flag          uint16 `type:"detail"`
	Value         int    `type:"value"`
	Worst         int    `type:"value"`
	Thresh        int    `type:"value"`
	Type          string `type:"detail"`
	Updated       string `type:"detail"`
	WhenFailed    string `type:"detail" name:"when_failed"`
	RawValue      int    `type:"value" name:"raw_value"`
	RawValueNotes string `type:"detail" name:"raw_value_notes"`
}

func newSmartAttribute(columns []string) *smartAttribute {
	attrib := &smartAttribute{}
	attrib.ID, _ = strconv.Atoi(columns[0])
	attrib.Name = columns[1]
	flag, _ := strconv.ParseUint(strings.Replace(columns[2], "0x", "", 1), 16, 16)
	attrib.Flag = uint16(flag)
	attrib.Value, _ = strconv.Atoi(columns[3])
	attrib.Worst, _ = strconv.Atoi(columns[4])
	attrib.Thresh, _ = strconv.Atoi(columns[5])
	attrib.Type = columns[6]
	attrib.Updated = columns[7]
	attrib.WhenFailed = columns[8]
	attrib.RawValue, _ = strconv.Atoi(columns[9])
	if len(columns) > 10 {
		attrib.RawValueNotes = columns[10]
	}
	return attrib
}

func (attr smartAttribute) getKey(useName bool, fieldName string) string {
	kparts := []string{fmt.Sprintf("%v", attr.ID)}
	if useName {
		kparts = append(kparts, attr.Name)
	}
	kparts = append(kparts, fieldName)
	return influxDBFieldNameFilterRgx.ReplaceAllString(strings.Join(kparts, "_"), "_")
}

func (attr smartAttribute) String(useNames, detailed bool) string {
	kv := make(map[string]string, 0)
	t := reflect.TypeOf(attr)
	v := reflect.ValueOf(attr)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		outputType := field.Tag.Get("type")

		if outputType == "value" || detailed && outputType == "detail" {
			fieldName := field.Tag.Get("name")
			key := ""
			if len(fieldName) > 0 {
				key = attr.getKey(useNames, fieldName)
			} else {
				key = attr.getKey(useNames, field.Name)
			}
			kv[key] = fmt.Sprintf("%v", v.Field(i).Interface())
		}
	}

	kvs := []string{}
	for k, v := range kv {
		kvs = append(kvs, fmt.Sprintf(`%v="%v"`, k, v))
	}

	return strings.Join(kvs, ",")
}

func parseAttributeList(out string) []*smartAttribute {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	var inTable bool
	attributes := make([]*smartAttribute, 0, len(lines)) // 7 is the usual header size
	for _, line := range lines {
		if strings.HasPrefix(line, "ID#") {
			inTable = true
			continue
		}
		if inTable {
			columns := columnsRgx.Split(strings.TrimSpace(line), 11)
			attributes = append(attributes, newSmartAttribute(columns))
		}
	}
	return attributes
}
