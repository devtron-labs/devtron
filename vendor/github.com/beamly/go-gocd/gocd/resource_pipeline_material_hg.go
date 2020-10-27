package gocd

import "errors"

func (mhg MaterialAttributesHg) equal(mhg2i MaterialAttribute) (bool, error) {
	var ok bool
	mhg2, ok := mhg2i.(MaterialAttributesHg)
	if !ok {
		return false, errors.New("can only compare with same material type")
	}
	urlsEqual := mhg.URL == mhg2.URL
	destEqual := mhg.Destination == mhg2.Destination

	return urlsEqual && destEqual, nil
}

// GenerateGeneric form (map[string]interface) of the material filter
func (mhg MaterialAttributesHg) GenerateGeneric() (ma map[string]interface{}) {
	ma = map[string]interface{}{
		"auto_update":   mhg.AutoUpdate,
		"destination":   mhg.Destination,
		"invert_filter": mhg.InvertFilter,
		"name":          mhg.Name,
		"url":           mhg.URL,
	}

	if f := mhg.Filter.GenerateGeneric(); f != nil {
		ma["filter"] = f
	}
	return
}

// HasFilter in this material attribute
func (mhg MaterialAttributesHg) HasFilter() bool {
	return true
}

// GetFilter from material attribute
func (mhg MaterialAttributesHg) GetFilter() *MaterialFilter {
	return mhg.Filter
}

// UnmarshallInterface from a JSON string to a MaterialAttributesHg struct
func unmarshallMaterialAttributesHg(mhg *MaterialAttributesHg, i map[string]interface{}) {
	for key, value := range i {
		if value == nil {
			continue
		}
		switch key {
		case "name":
			mhg.Name = value.(string)
		case "url":
			mhg.URL = value.(string)
		case "destination":
			mhg.Destination = value.(string)
		case "invert_filter":
			mhg.InvertFilter = value.(bool)
		case "auto_update":
			mhg.AutoUpdate = value.(bool)
		case "filter":
			mhg.Filter = unmarshallMaterialFilter(value.(map[string]interface{}))
		}
	}
}
