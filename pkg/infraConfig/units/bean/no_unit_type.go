package bean

import serviceBean "github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"

type NoUnitStr string

const (
	NoUnit NoUnitStr = ""
)

// GetUnitSuffix returns the UnitSuffix for NoUnit (just return 20 as it represents no unit)
func (noUnitStr NoUnitStr) GetUnitSuffix() UnitType {
	switch noUnitStr {
	case NoUnit:
		return NoSuffix
	default:
		return NoSuffix
	}
}

func (noUnitStr NoUnitStr) GetUnit() (serviceBean.Unit, bool) {
	noUnits := GetNoUnit()
	noUnit, exists := noUnits[noUnitStr]
	return noUnit, exists
}

func (noUnitStr NoUnitStr) String() string {
	return string(noUnitStr)
}

func GetNoUnit() map[NoUnitStr]serviceBean.Unit {
	return map[NoUnitStr]serviceBean.Unit{
		NoUnit: {
			Name:             "NoUnit",
			ConversionFactor: 0,
		},
	}
}
