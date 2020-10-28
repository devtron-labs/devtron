package gocd

import (
	"errors"
)

func (mad MaterialAttributesDependency) equal(mad2i MaterialAttribute) (bool, error) {
	var ok bool
	mad2, ok := mad2i.(MaterialAttributesDependency)
	if !ok {
		return false, errors.New("can only compare with same material type")
	}
	return mad.Pipeline == mad2.Pipeline &&
			mad.Stage == mad2.Stage,
		nil
}

// GenerateGeneric form (map[string]interface) of the material filter
func (mad MaterialAttributesDependency) GenerateGeneric() (ma map[string]interface{}) {
	ma = make(map[string]interface{})
	if mad.Name != "" {
		ma["name"] = mad.Name
	}

	if mad.Pipeline != "" {
		ma["pipeline"] = mad.Pipeline
	}

	if mad.Stage != "" {
		ma["stage"] = mad.Stage
	}

	if mad.AutoUpdate {
		ma["auto_update"] = mad.AutoUpdate
	}
	return
}

// HasFilter in this material attribute
func (mad MaterialAttributesDependency) HasFilter() bool {
	return false
}

// GetFilter from material attribute
func (mad MaterialAttributesDependency) GetFilter() *MaterialFilter {
	return nil
}

// UnmarshallInterface for a MaterialAttribute struct to be turned into a json string
func unmarshallMaterialAttributesDependency(mad *MaterialAttributesDependency, i map[string]interface{}) {
	for key, value := range i {
		if value == nil {
			continue
		}
		switch key {
		case "name":
			mad.Name = value.(string)
		case "pipeline":
			mad.Pipeline = value.(string)
		case "stage":
			mad.Stage = value.(string)
		case "auto_update":
			mad.AutoUpdate = value.(bool)
		}
	}
}
