package gocd

import "errors"

func (mapp MaterialAttributesPlugin) equal(mapp2i MaterialAttribute) (bool, error) {
	var ok bool
	mapp2, ok := mapp2i.(MaterialAttributesPlugin)
	if !ok {
		return false, errors.New("can only compare with same material type")
	}

	return mapp.Ref == mapp2.Ref &&
			mapp.Destination == mapp2.Destination,
		nil
}

// GenerateGeneric form (map[string]interface) of the material filter
func (mapp MaterialAttributesPlugin) GenerateGeneric() (ma map[string]interface{}) {
	ma = make(map[string]interface{})
	return
}

// HasFilter in this material attribute
func (mapp MaterialAttributesPlugin) HasFilter() bool {
	return true
}

// GetFilter from material attribute
func (mapp MaterialAttributesPlugin) GetFilter() *MaterialFilter {
	return mapp.Filter
}

// UnmarshallInterface from a JSON string to a MaterialAttributesPlugin struct
func unmarshallMaterialAttributesPlugin(mapp *MaterialAttributesPlugin, i map[string]interface{}) {
	for key, value := range i {
		if value == nil {
			continue
		}
		switch key {
		case "ref":
			mapp.Ref = value.(string)
		case "destination":
			mapp.Destination = value.(string)
		case "filter":
			mapp.Filter = unmarshallMaterialFilter(value.(map[string]interface{}))
		case "invert_filter":
			mapp.InvertFilter = value.(bool)
		}
	}
}
