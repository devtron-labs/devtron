package gocd

import "errors"

func (mapk MaterialAttributesPackage) equal(mapk2i MaterialAttribute) (bool, error) {
	var ok bool
	mapk2, ok := mapk2i.(MaterialAttributesPackage)
	if !ok {
		return false, errors.New("can only compare with same material type")
	}

	return mapk.Ref == mapk2.Ref, nil
}

// GenerateGeneric form (map[string]interface) of the material filter
func (mapk MaterialAttributesPackage) GenerateGeneric() (ma map[string]interface{}) {
	ma = make(map[string]interface{})
	return
}

// HasFilter in this material attribute
func (mapk MaterialAttributesPackage) HasFilter() bool {
	return false
}

// GetFilter from material attribute
func (mapk MaterialAttributesPackage) GetFilter() *MaterialFilter {
	return nil
}

// UnmarshallInterface from a JSON string to a MaterialAttributesPackage struct
func unmarshallMaterialAttributesPackage(mapk *MaterialAttributesPackage, i map[string]interface{}) {
	for key, value := range i {
		if value == nil {
			continue
		}
		switch key {
		case "ref":
			mapk.Ref = value.(string)
		}
	}
}
