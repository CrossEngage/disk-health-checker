package main

import (
	"regexp"
	"strconv"
	"strings"
)

func parseSMARTCtlScan(out string) []string {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	devices := make([]string, 0, len(lines))
	for _, line := range lines {
		tmp := strings.Split(line, "#")
		devices = append(devices, strings.TrimSpace(tmp[0]))
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
}

var (
	separateSectorSizesRgx = regexp.MustCompile(`^(\d+) bytes logical, (\d+) bytes physical$`)
	sameSectorSizesRgx     = regexp.MustCompile(`^(\d+) bytes logical\/physical$`)
	userCapacityRgx        = regexp.MustCompile(`^([,\d]+) bytes \[.+?\]$`)
	rpmRgx                 = regexp.MustCompile(`^([\d]+) rpm$`)
	bytesRgx               = regexp.MustCompile(`^(\d+) bytes$`)
)

func parseSMARTCtlInfo(out string) *smartCtlInfo {
	info := &smartCtlInfo{}
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
		}
	}
	return info
}
