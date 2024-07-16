package speedtest

import (
	"strconv"
)

type UnitType int

// IEC and SI
const (
	UnitTypeDecimalBits  = UnitType(iota) // auto scaled
	UnitTypeDecimalBytes                  // auto scaled
	UnitTypeBinaryBits                    // auto scaled
	UnitTypeBinaryBytes                   // auto scaled
	UnitTypeDefaultMbps                   // fixed
)

var (
	DecimalBitsUnits  = []string{"bps", "Kbps", "Mbps", "Gbps"}
	DecimalBytesUnits = []string{"B/s", "KB/s", "MB/s", "GB/s"}
	BinaryBitsUnits   = []string{"Kibps", "KiMbps", "KiGbps"}
	BinaryBytesUnits  = []string{"KiB/s", "MiB/s", "GiB/s"}
)

var unitMaps = map[UnitType][]string{
	UnitTypeDecimalBits:  DecimalBitsUnits,
	UnitTypeDecimalBytes: DecimalBytesUnits,
	UnitTypeBinaryBits:   BinaryBitsUnits,
	UnitTypeBinaryBytes:  BinaryBytesUnits,
}

const (
	B  = 1.0
	KB = 1000 * B
	MB = 1000 * KB
	GB = 1000 * MB

	IB  = 1
	KiB = 1024 * IB
	MiB = 1024 * KiB
	GiB = 1024 * MiB
)

type ByteRate float64

var globalByteRateUnit UnitType

func (r ByteRate) String() string {
	if r == 0 {
		return "0.00 Mbps"
	}
	if r == -1 {
		return "N/A"
	}
	if globalByteRateUnit != UnitTypeDefaultMbps {
		return r.Byte(globalByteRateUnit)
	}
	return strconv.FormatFloat(float64(r/125000.0), 'f', 2, 64) + " Mbps"
}

// SetUnit Set global output units
func SetUnit(unit UnitType) {
	globalByteRateUnit = unit
}

func (r ByteRate) Mbps() float64 {
	return float64(r) / 125000.0
}

func (r ByteRate) Gbps() float64 {
	return float64(r) / 125000000.0
}

// Byte Specifies the format output byte rate
func (r ByteRate) Byte(formatType UnitType) string {
	if r == 0 {
		return "0.00 Mbps"
	}
	if r == -1 {
		return "N/A"
	}
	return format(float64(r), formatType)
}

func format(byteRate float64, i UnitType) string {
	val := byteRate
	if i%2 == 0 {
		val *= 8
	}
	if i < UnitTypeBinaryBits {
		switch {
		case byteRate >= GB:
			return strconv.FormatFloat(val/GB, 'f', 2, 64) + " " + unitMaps[i][3]
		case byteRate >= MB:
			return strconv.FormatFloat(val/MB, 'f', 2, 64) + " " + unitMaps[i][2]
		case byteRate >= KB:
			return strconv.FormatFloat(val/KB, 'f', 2, 64) + " " + unitMaps[i][1]
		default:
			return strconv.FormatFloat(val/B, 'f', 2, 64) + " " + unitMaps[i][0]
		}
	}
	switch {
	case byteRate >= GiB:
		return strconv.FormatFloat(val/GiB, 'f', 2, 64) + " " + unitMaps[i][2]
	case byteRate >= MiB:
		return strconv.FormatFloat(val/MiB, 'f', 2, 64) + " " + unitMaps[i][1]
	default:
		return strconv.FormatFloat(val/KiB, 'f', 2, 64) + " " + unitMaps[i][0]
	}
}
