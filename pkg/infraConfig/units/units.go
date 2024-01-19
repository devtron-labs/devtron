package units

import (
	"k8s.io/apimachinery/pkg/api/resource"
	"strings"
)

// memory units
// Ei, Pi, Ti, Gi, Mi, Ki
// E, P, T, G, M, k

type UnitSuffix int

const (
	Byte   UnitSuffix = 0
	KiByte UnitSuffix = 1
	MiByte UnitSuffix = 2
	GiByte UnitSuffix = 3
	TiByte UnitSuffix = 4
	PiByte UnitSuffix = 5
	EiByte UnitSuffix = 6
	K      UnitSuffix = 7
	M      UnitSuffix = 8
	G      UnitSuffix = 9
	T      UnitSuffix = 10
	P      UnitSuffix = 11
	E      UnitSuffix = 12
	Core   UnitSuffix = 13
	Milli  UnitSuffix = 14
	Second UnitSuffix = 15
	Minute UnitSuffix = 16
	Hour   UnitSuffix = 17
)

type CPUUnitStr string

func GetCPUUnitStr(cpuUnit UnitSuffix) CPUUnitStr {
	switch cpuUnit {
	case Core:
		return CORE
	case Milli:
		return MILLI
	default:
		return CORE
	}
}

func GetCPUUnit(cpuUnitStr CPUUnitStr) UnitSuffix {
	switch cpuUnitStr {
	case CORE:
		return Core
	case MILLI:
		return Milli
	default:
		return Core
	}
}

const (
	CORE  CPUUnitStr = ""
	MILLI CPUUnitStr = "m"
)

type MemoryUnitStr string

const (
	BYTE   MemoryUnitStr = ""
	KIBYTE MemoryUnitStr = "Ki"
	MIBYTE MemoryUnitStr = "Mi"
	GIBYTE MemoryUnitStr = "Gi"
	TIBYTE MemoryUnitStr = "Ti"
	PIBYTE MemoryUnitStr = "Pi"
	EIBYTE MemoryUnitStr = "Ei"
	KBYTE  MemoryUnitStr = "k"
	MBYTE  MemoryUnitStr = "M"
	GBYTE  MemoryUnitStr = "G"
	TBYTE  MemoryUnitStr = "T"
	PBYTE  MemoryUnitStr = "P"
	EBYTE  MemoryUnitStr = "E"
)

func GetMemoryUnitStr(memoryUnit UnitSuffix) MemoryUnitStr {
	switch memoryUnit {
	case Byte:
		return BYTE
	case KiByte:
		return KIBYTE
	case MiByte:
		return MIBYTE
	case GiByte:
		return GIBYTE
	case TiByte:
		return TIBYTE
	case PiByte:
		return PIBYTE
	case EiByte:
		return EIBYTE
	case K:
		return KBYTE
	case M:
		return MBYTE
	case G:
		return GBYTE
	case T:
		return TBYTE
	case P:
		return PBYTE
	case E:
		return EBYTE
	default:
		return BYTE
	}
}

func GetMemoryUnit(memoryUnitStr MemoryUnitStr) UnitSuffix {
	switch memoryUnitStr {
	case BYTE:
		return Byte
	case KIBYTE:
		return KiByte
	case MIBYTE:
		return MiByte
	case GIBYTE:
		return GiByte
	case TIBYTE:
		return TiByte
	case PIBYTE:
		return PiByte
	case EIBYTE:
		return EiByte
	case KBYTE:
		return K
	case MBYTE:
		return M
	case GBYTE:
		return G
	case TBYTE:
		return T
	case PBYTE:
		return P
	case EBYTE:
		return E
	default:
		return Byte
	}
}

type TimeUnitStr string

const (
	SecondStr TimeUnitStr = "s"
	MinuteStr TimeUnitStr = "m"
	HourStr   TimeUnitStr = "h"
)

func GetTimeUnitStr(timeUnit UnitSuffix) TimeUnitStr {
	switch timeUnit {
	case Second:
		return SecondStr
	case Minute:
		return MinuteStr
	case Hour:
		return HourStr
	default:
		return SecondStr
	}
}

func GetTimeUnit(timeUnitStr TimeUnitStr) UnitSuffix {
	switch timeUnitStr {
	case SecondStr:
		return Second
	case MinuteStr:
		return Minute
	case HourStr:
		return Hour
	default:
		return Second
	}
}

type Units struct {
	cpuUnits    []Unit
	memoryUnits []Unit
	timeUnits   []Unit
}

func NewUnits() *Units {
	cpuUnits := [2]Unit{
		{
			Name:             string(MILLI),
			ConversionFactor: 1000,
		},
		{
			Name:             string(CORE),
			ConversionFactor: 1,
		},
	}

	memoryUnits := [13]Unit{
		{
			Name:             string(BYTE),
			ConversionFactor: 1,
		},
		{
			Name:             string(KBYTE),
			ConversionFactor: 1 / 1000,
		},
		{
			Name:             string(MBYTE),
			ConversionFactor: 1 / 1000000,
		},
		{
			Name:             string(GBYTE),
			ConversionFactor: 1 / 1000000000,
		},
		{
			Name:             string(TBYTE),
			ConversionFactor: 1 / 1000000000000,
		},
		{
			Name:             string(PBYTE),
			ConversionFactor: 1 / 1000000000000000,
		},
		{
			Name:             string(EBYTE),
			ConversionFactor: 1 / 1000000000000000000,
		},
		{
			Name:             string(KIBYTE),
			ConversionFactor: 1 / (1024),
		},
		{
			Name:             string(MIBYTE),
			ConversionFactor: 1 / (1024 * 1024),
		},
		{
			Name:             string(GIBYTE),
			ConversionFactor: 1 / (1024 * 1024 * 1024),
		},
		{
			Name:             string(TIBYTE),
			ConversionFactor: 1 / (1024 * 1024 * 1024 * 1024),
		},
		{
			Name:             string(PIBYTE),
			ConversionFactor: 1 / (1024 * 1024 * 1024 * 1024 * 1024),
		},
		{
			Name:             string(EIBYTE),
			ConversionFactor: 1 / (1024 * 1024 * 1024 * 1024 * 1024 * 1024),
		},
	}

	timeUnits := [3]Unit{
		{
			Name:             "s",
			ConversionFactor: 1,
		},
		{
			Name:             "m",
			ConversionFactor: 1 / 60,
		},
		{
			Name:             "h",
			ConversionFactor: 1 / 3600,
		},
	}
	return &Units{

		cpuUnits:    cpuUnits[:],
		memoryUnits: memoryUnits[:],
		timeUnits:   timeUnits[:],
	}
}

func (u *Units) GetCpuUnits() []Unit {
	return u.cpuUnits
}
func (u *Units) GetMemoryUnits() []Unit {
	return u.memoryUnits
}

func (u *Units) GetTimeUnits() []Unit {
	return u.timeUnits
}

// Unit represents unit of a configuration
type Unit struct {
	// Name is unit name
	Name string `json:"name"`
	// ConversionFactor is used to convert this unit to the base unit
	// if ConversionFactor is 1, then this is the base unit
	ConversionFactor float64 `json:"conversionFactor"`
}

// ParseQuantityString is a fast scanner for quantity values.
// this parsing is only for cpu and mem resources
func ParseQuantityString(str string) (positive bool, value, num, denom, suffix string, err error) {
	positive = true
	pos := 0
	end := len(str)

	// handle leading sign
	if pos < end {
		switch str[0] {
		case '-':
			positive = false
			pos++
		case '+':
			pos++
		}
	}

	// strip leading zeros
Zeroes:
	for i := pos; ; i++ {
		if i >= end {
			num = "0"
			value = num
			return
		}
		switch str[i] {
		case '0':
			pos++
		default:
			break Zeroes
		}
	}

	// extract the numerator
Num:
	for i := pos; ; i++ {
		if i >= end {
			num = str[pos:end]
			value = str[0:end]
			return
		}
		switch str[i] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		default:
			num = str[pos:i]
			pos = i
			break Num
		}
	}

	// if we stripped all numerator positions, always return 0
	if len(num) == 0 {
		num = "0"
	}

	// handle a denominator
	if pos < end && str[pos] == '.' {
		pos++
	Denom:
		for i := pos; ; i++ {
			if i >= end {
				denom = str[pos:end]
				value = str[0:end]
				return
			}
			switch str[i] {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			default:
				denom = str[pos:i]
				pos = i
				break Denom
			}
		}
		// TODO: we currently allow 1.G, but we may not want to in the future.
		// if len(denom) == 0 {
		// 	err = ErrFormatWrong
		// 	return
		// }
	}
	value = str[0:pos]

	// grab the elements of the suffix
	suffixStart := pos
	for i := pos; ; i++ {
		if i >= end {
			suffix = str[suffixStart:end]
			return
		}
		if !strings.ContainsAny(str[i:i+1], "sheEinumkKMGTP") {
			pos = i
			break
		}
	}
	if pos < end {
		switch str[pos] {
		case '-', '+':
			pos++
		}
	}
Suffix:
	for i := pos; ; i++ {
		if i >= end {
			suffix = str[suffixStart:end]
			return
		}
		switch str[i] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		default:
			break Suffix
		}
	}
	// we encountered a non decimal in the Suffix loop, but the last character
	// was not a valid exponent
	err = resource.ErrFormatWrong
	return
}
