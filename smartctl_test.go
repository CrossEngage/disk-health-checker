package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSMARTDevices(t *testing.T) {
	var scanOutput = `
/dev/sda -d scsi # /dev/sda, SCSI device
/dev/sdb -d scsi # /dev/sdb, SCSI device
`
	devices := parseSMARTCtlScan(scanOutput)
	assert.NotEmpty(t, devices)
	assert.Equal(t, "/dev/sda", devices[0].Path)
	assert.Equal(t, "auto", devices[0].Type)
	assert.Equal(t, "/dev/sdb", devices[1].Path)
	assert.Equal(t, "auto", devices[1].Type)
}

func TestGetSMARTDevicesOnHWRAID(t *testing.T) {
	var scanHWRAIDOutput = `
/dev/sda -d scsi # /dev/sda, SCSI device
/dev/bus/4 -d megaraid,14 # /dev/bus/4 [megaraid_disk_14], SCSI device
/dev/bus/4 -d megaraid,15 # /dev/bus/4 [megaraid_disk_15], SCSI device
`
	devices := parseSMARTCtlScan(scanHWRAIDOutput)
	assert.NotEmpty(t, devices)
	assert.Equal(t, "/dev/sda", devices[0].Path)
	assert.Equal(t, "auto", devices[0].Type)
	assert.Equal(t, "/dev/bus/4", devices[1].Path)
	assert.Equal(t, "megaraid,14", devices[1].Type)
	assert.Equal(t, "/dev/bus/4", devices[2].Path)
	assert.Equal(t, "megaraid,15", devices[2].Type)
}

func TestParseSMARTCtlInfo(t *testing.T) {
	var diskInfoOutput = `
smartctl 6.2 2013-07-26 r3841 [x86_64-linux-4.4.0-47-generic] (local build)
Copyright (C) 2002-13, Bruce Allen, Christian Franke, www.smartmontools.org

=== START OF INFORMATION SECTION ===
Device Model:     HGST HDN724040ALE640
Serial Number:    PK2338P4H4XPXC
LU WWN Device Id: 5 000cca 249d054c0
Firmware Version: MJAOA5E0
User Capacity:    4,000,787,030,016 bytes [4.00 TB]
Sector Sizes:     512 bytes logical, 4096 bytes physical
Rotation Rate:    7200 rpm
Device is:        Not in smartctl database [for details use: -P showall]
ATA Version is:   ATA8-ACS T13/1699-D revision 4
SATA Version is:  SATA 3.0, 6.0 Gb/s (current: 6.0 Gb/s)
Local Time is:    Tue Jun 27 10:21:29 2017 CEST
SMART support is: Available - device has SMART capability.
SMART support is: Enabled

=== START OF READ SMART DATA SECTION ===
SMART overall-health self-assessment test result: PASSED

`
	info := parseSMARTCtlInfo(diskInfoOutput)
	assert.Equal(t, "HGST HDN724040ALE640", info.DeviceModel)
	assert.Equal(t, "PK2338P4H4XPXC", info.SerialNumber)
	assert.Equal(t, "5 000cca 249d054c0", info.LUWWNDeviceID)
	assert.Equal(t, "MJAOA5E0", info.FirmwareVersion)
	assert.Equal(t, "4,000,787,030,016 bytes [4.00 TB]", info.UserCapacity)
	assert.Equal(t, int64(4000787030016), info.UserCapacityBytes)
	assert.Equal(t, "512 bytes logical, 4096 bytes physical", info.SectorSizes)
	assert.Equal(t, 512, info.LogicalSectorSizes)
	assert.Equal(t, 4096, info.PhysicalSectorSizes)
	assert.Equal(t, "7200 rpm", info.RotationRate)
	assert.Equal(t, 7200, info.RotationRateRPM)
	assert.Equal(t, "ATA8-ACS T13/1699-D revision 4", info.ATAVersion)
	assert.Equal(t, "SATA 3.0, 6.0 Gb/s (current: 6.0 Gb/s)", info.SATAVersion)
	assert.Equal(t, "Enabled", info.SMARTSupportIs)
	assert.Equal(t, true, info.SMARTSupport)
}

func TestParseSSDSMARTCtlInfo(t *testing.T) {
	var ssdDiskInfoOutput = `
smartctl 6.2 2013-07-26 r3841 [x86_64-linux-3.16.0-77-generic] (local build)
Copyright (C) 2002-13, Bruce Allen, Christian Franke, www.smartmontools.org

=== START OF INFORMATION SECTION ===
Device Model:     INTEL SSDSC2BW480H6
Serial Number:    CVTR527201W1480EGN
LU WWN Device Id: 5 5cd2e4 14c8c2337
Firmware Version: RG20
User Capacity:    480,103,981,056 bytes [480 GB]
Sector Size:      512 bytes logical/physical
Rotation Rate:    Solid State Device
Device is:        Not in smartctl database [for details use: -P showall]
ATA Version is:   ACS-3 (minor revision not indicated)
SATA Version is:  SATA >3.1, 6.0 Gb/s (current: 6.0 Gb/s)
Local Time is:    Tue Jun 27 10:51:17 2017 CEST
SMART support is: Available - device has SMART capability.
SMART support is: Enabled

=== START OF READ SMART DATA SECTION ===
SMART overall-health self-assessment test result: PASSED

`
	info := parseSMARTCtlInfo(ssdDiskInfoOutput)
	assert.Equal(t, "INTEL SSDSC2BW480H6", info.DeviceModel)
	assert.Equal(t, "CVTR527201W1480EGN", info.SerialNumber)
	assert.Equal(t, "5 5cd2e4 14c8c2337", info.LUWWNDeviceID)
	assert.Equal(t, "RG20", info.FirmwareVersion)
	assert.Equal(t, "480,103,981,056 bytes [480 GB]", info.UserCapacity)
	assert.Equal(t, int64(480103981056), info.UserCapacityBytes)
	assert.Equal(t, "512 bytes logical/physical", info.SectorSizes)
	assert.Equal(t, 512, info.LogicalSectorSizes)
	assert.Equal(t, 512, info.PhysicalSectorSizes)
	assert.Equal(t, "Solid State Device", info.RotationRate)
	assert.Equal(t, true, info.IsSSD)
	assert.Equal(t, 0, info.RotationRateRPM)
	assert.Equal(t, "ACS-3 (minor revision not indicated)", info.ATAVersion)
	assert.Equal(t, "SATA >3.1, 6.0 Gb/s (current: 6.0 Gb/s)", info.SATAVersion)
	assert.Equal(t, "Enabled", info.SMARTSupportIs)
	assert.Equal(t, true, info.SMARTSupport)
}

func TestParseRaidVolumeSMARTCtlInfo(t *testing.T) {
	var raidVolumeInfoOutput = `
smartctl 6.2 2013-07-26 r3841 [x86_64-linux-3.10.0-327.10.1.el7.x86_64] (local build)
Copyright (C) 2002-13, Bruce Allen, Christian Franke, www.smartmontools.org

=== START OF INFORMATION SECTION ===
Vendor:               LSI
Product:              Logical Volume
Revision:             3000
User Capacity:        298,999,349,248 bytes [298 GB]
Logical block size:   512 bytes
Logical Unit id:      0x600508e0000000006b402971cbb20d0f
Device type:          disk
Local Time is:        Tue Jun 27 11:37:47 2017 CEST
SMART support is:     Unavailable - device lacks SMART capability.

=== START OF READ SMART DATA SECTION ===

`
	info := parseSMARTCtlInfo(raidVolumeInfoOutput)
	assert.Equal(t, "LSI", info.Vendor)
	assert.Equal(t, "Logical Volume", info.Product)
	assert.Equal(t, "3000", info.Revision)
	assert.Equal(t, "298,999,349,248 bytes [298 GB]", info.UserCapacity)
	assert.Equal(t, int64(298999349248), info.UserCapacityBytes)
	assert.Equal(t, "512 bytes", info.LogicalBlockSize)
	assert.Equal(t, 512, info.LogicalBlockSizeBytes)
	assert.Equal(t, "0x600508e0000000006b402971cbb20d0f", info.LogicalUnitID)
	assert.Equal(t, "disk", info.DeviceType)
	assert.Equal(t, "Unavailable - device lacks SMART capability.", info.SMARTSupportIs)
	assert.Equal(t, false, info.SMARTSupport)
	assert.Equal(t, "UNSUPPORTED", info.Health)
	assert.Equal(t, false, info.Healthy)
}

func TestParseDiskAttributesOutput(t *testing.T) {
	var diskAttributesOutput = `
smartctl 6.2 2013-07-26 r3841 [x86_64-linux-4.4.0-47-generic] (local build)
Copyright (C) 2002-13, Bruce Allen, Christian Franke, www.smartmontools.org

=== START OF READ SMART DATA SECTION ===
SMART Attributes Data Structure revision number: 16
Vendor Specific SMART Attributes with Thresholds:
ID# ATTRIBUTE_NAME          FLAG     VALUE WORST THRESH TYPE      UPDATED  WHEN_FAILED RAW_VALUE
  1 Raw_Read_Error_Rate     0x000b   100   100   016    Pre-fail  Always       -       0
  2 Throughput_Performance  0x0005   136   136   054    Pre-fail  Offline      -       80
  3 Spin_Up_Time            0x0007   142   142   024    Pre-fail  Always       -       611 (Average 480)
  4 Start_Stop_Count        0x0012   100   100   000    Old_age   Always       -       11
  5 Reallocated_Sector_Ct   0x0033   100   100   005    Pre-fail  Always       -       0
  7 Seek_Error_Rate         0x000b   100   100   067    Pre-fail  Always       -       0
  8 Seek_Time_Performance   0x0005   121   121   020    Pre-fail  Offline      -       34
  9 Power_On_Hours          0x0012   098   098   000    Old_age   Always       -       14605
 10 Spin_Retry_Count        0x0013   100   100   060    Pre-fail  Always       -       0
 12 Power_Cycle_Count       0x0032   100   100   000    Old_age   Always       -       11
192 Power-Off_Retract_Count 0x0032   100   100   000    Old_age   Always       -       132
193 Load_Cycle_Count        0x0012   100   100   000    Old_age   Always       -       132
194 Temperature_Celsius     0x0002   157   157   000    Old_age   Always       -       38 (Min/Max 24/45)
196 Reallocated_Event_Count 0x0032   100   100   000    Old_age   Always       -       0
197 Current_Pending_Sector  0x0022   100   100   000    Old_age   Always       -       0
198 Offline_Uncorrectable   0x0008   100   100   000    Old_age   Offline      -       0
199 UDMA_CRC_Error_Count    0x000a   200   200   000    Old_age   Always       -       0

`
	attributes := parseAttributeList(diskAttributesOutput)
	assert.NotNil(t, attributes)
	assert.Len(t, attributes, 17)
	assert.Equal(t, 194, attributes[12].ID)
	assert.Equal(t, "Temperature_Celsius", attributes[12].Name)
	assert.Equal(t, uint16(0x0002), attributes[12].Flag)
	assert.Equal(t, 157, attributes[12].Value)
	assert.Equal(t, 157, attributes[12].Worst)
	assert.Equal(t, 0, attributes[12].Thresh)
	assert.Equal(t, "Old_age", attributes[12].Type)
	assert.Equal(t, "Always", attributes[12].Updated)
	assert.Equal(t, "-", attributes[12].WhenFailed)
	assert.Equal(t, 38, attributes[12].RawValue)
	assert.Equal(t, "(Min/Max 24/45)", attributes[12].RawValueNotes)
}

func TestParseSSDDiskAttributesOutput(t *testing.T) {
	var ssdDiskAttributesOutput = `
smartctl 6.2 2013-07-26 r3841 [x86_64-linux-3.16.0-77-generic] (local build)
Copyright (C) 2002-13, Bruce Allen, Christian Franke, www.smartmontools.org

=== START OF READ SMART DATA SECTION ===
SMART Attributes Data Structure revision number: 10
Vendor Specific SMART Attributes with Thresholds:
ID# ATTRIBUTE_NAME          FLAG     VALUE WORST THRESH TYPE      UPDATED  WHEN_FAILED RAW_VALUE
  5 Reallocated_Sector_Ct   0x0032   100   100   000    Old_age   Always       -       0
  9 Power_On_Hours          0x0032   100   100   000    Old_age   Always       -       5082
 12 Power_Cycle_Count       0x0032   100   100   000    Old_age   Always       -       7
170 Unknown_Attribute       0x0033   097   100   010    Pre-fail  Always       -       0
171 Unknown_Attribute       0x0032   100   100   000    Old_age   Always       -       0
172 Unknown_Attribute       0x0032   100   100   000    Old_age   Always       -       0
174 Unknown_Attribute       0x0032   100   100   000    Old_age   Always       -       1
183 Runtime_Bad_Block       0x0032   100   100   000    Old_age   Always       -       2
184 End-to-End_Error        0x0033   100   100   090    Pre-fail  Always       -       0
187 Reported_Uncorrect      0x0032   100   100   000    Old_age   Always       -       0
190 Airflow_Temperature_Cel 0x0032   027   100   000    Old_age   Always       -       27 (Min/Max 25/42)
192 Power-Off_Retract_Count 0x0032   100   100   000    Old_age   Always       -       1
199 UDMA_CRC_Error_Count    0x0032   100   100   000    Old_age   Always       -       0
225 Unknown_SSD_Attribute   0x0032   100   100   000    Old_age   Always       -       1726241
226 Unknown_SSD_Attribute   0x0032   100   100   000    Old_age   Always       -       65535
227 Unknown_SSD_Attribute   0x0032   100   100   000    Old_age   Always       -       24
228 Power-off_Retract_Count 0x0032   100   100   000    Old_age   Always       -       65535
232 Available_Reservd_Space 0x0033   097   100   010    Pre-fail  Always       -       0
233 Media_Wearout_Indicator 0x0032   074   100   000    Old_age   Always       -       0
241 Total_LBAs_Written      0x0032   100   100   000    Old_age   Always       -       1726241
242 Total_LBAs_Read         0x0032   100   100   000    Old_age   Always       -       566621
249 Unknown_Attribute       0x0032   100   100   000    Old_age   Always       -       200932

`
	attributes := parseAttributeList(ssdDiskAttributesOutput)
	assert.NotNil(t, attributes)
	assert.Len(t, attributes, 22)
	assert.Equal(t, 199, attributes[12].ID)
	assert.Equal(t, "UDMA_CRC_Error_Count", attributes[12].Name)
	assert.Equal(t, uint16(0x0032), attributes[12].Flag)
	assert.Equal(t, 100, attributes[12].Value)
	assert.Equal(t, 100, attributes[12].Worst)
	assert.Equal(t, 0, attributes[12].Thresh)
	assert.Equal(t, "Old_age", attributes[12].Type)
	assert.Equal(t, "Always", attributes[12].Updated)
	assert.Equal(t, "-", attributes[12].WhenFailed)
	assert.Equal(t, 0, attributes[12].RawValue)
	assert.Equal(t, "", attributes[12].RawValueNotes)
}

func TestParseHWRAIDDiskAttributesOutput(t *testing.T) {
	var hwRAIDDiskAttributesOutput = `
smartctl 6.6 2016-05-31 r4324 [x86_64-linux-4.4.35-2-pve] (local build)
Copyright (C) 2002-16, Bruce Allen, Christian Franke, www.smartmontools.org

=== START OF READ SMART DATA SECTION ===
SMART Attributes Data Structure revision number: 10
Vendor Specific SMART Attributes with Thresholds:
ID# ATTRIBUTE_NAME          FLAG     VALUE WORST THRESH TYPE      UPDATED  WHEN_FAILED RAW_VALUE
  1 Raw_Read_Error_Rate     0x000f   117   099   006    Pre-fail  Always       -       151524336
  3 Spin_Up_Time            0x0003   097   097   000    Pre-fail  Always       -       0
  4 Start_Stop_Count        0x0032   100   100   020    Old_age   Always       -       32
  5 Reallocated_Sector_Ct   0x0033   100   100   010    Pre-fail  Always       -       0
  7 Seek_Error_Rate         0x000f   080   060   030    Pre-fail  Always       -       107999872
  9 Power_On_Hours          0x0032   063   063   000    Old_age   Always       -       33211
 10 Spin_Retry_Count        0x0013   100   100   097    Pre-fail  Always       -       0
 12 Power_Cycle_Count       0x0032   100   100   020    Old_age   Always       -       31
183 Runtime_Bad_Block       0x0032   100   100   000    Old_age   Always       -       0
184 End-to-End_Error        0x0032   100   100   099    Old_age   Always       -       0
187 Reported_Uncorrect      0x0032   100   100   000    Old_age   Always       -       0
188 Command_Timeout         0x0032   100   100   000    Old_age   Always       -       0 0 0
189 High_Fly_Writes         0x003a   100   100   000    Old_age   Always       -       0
190 Airflow_Temperature_Cel 0x0022   074   059   045    Old_age   Always       -       26 (Min/Max 23/28)
191 G-Sense_Error_Rate      0x0032   100   100   000    Old_age   Always       -       0
192 Power-Off_Retract_Count 0x0032   100   100   000    Old_age   Always       -       31
193 Load_Cycle_Count        0x0032   001   001   000    Old_age   Always       -       205541
194 Temperature_Celsius     0x0022   026   041   000    Old_age   Always       -       26 (0 18 0 0 0)
197 Current_Pending_Sector  0x0012   100   100   000    Old_age   Always       -       0
198 Offline_Uncorrectable   0x0010   100   100   000    Old_age   Offline      -       0
199 UDMA_CRC_Error_Count    0x003e   200   200   000    Old_age   Always       -       0
240 Head_Flying_Hours       0x0000   100   253   000    Old_age   Offline      -       31467h+00m+35.937s
241 Total_LBAs_Written      0x0000   100   253   000    Old_age   Offline      -       20081879990
242 Total_LBAs_Read         0x0000   100   253   000    Old_age   Offline      -       54631063470

`
	attributes := parseAttributeList(hwRAIDDiskAttributesOutput)
	assert.NotNil(t, attributes)
	assert.Len(t, attributes, 24)
	assert.Equal(t, 189, attributes[12].ID)
	assert.Equal(t, "High_Fly_Writes", attributes[12].Name)
	assert.Equal(t, uint16(0x003a), attributes[12].Flag)
	assert.Equal(t, 100, attributes[12].Value)
	assert.Equal(t, 100, attributes[12].Worst)
	assert.Equal(t, 0, attributes[12].Thresh)
	assert.Equal(t, "Old_age", attributes[12].Type)
	assert.Equal(t, "Always", attributes[12].Updated)
	assert.Equal(t, "-", attributes[12].WhenFailed)
	assert.Equal(t, 0, attributes[12].RawValue)
	assert.Equal(t, "", attributes[12].RawValueNotes)
}
