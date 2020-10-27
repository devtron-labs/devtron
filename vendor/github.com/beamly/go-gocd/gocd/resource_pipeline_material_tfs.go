package gocd

import (
	"errors"
)

func (mtfs MaterialAttributesTfs) equal(mtfs2i MaterialAttribute) (bool, error) {
	var ok bool
	mtfs2, ok := mtfs2i.(MaterialAttributesTfs)
	if !ok {
		return false, errors.New("can only compare with same material type")
	}
	namesEqual := mtfs.Name == mtfs2.Name
	urlsEqual := mtfs.URL == mtfs2.URL
	projectsEqual := mtfs.ProjectPath == mtfs2.ProjectPath
	domainsEqual := mtfs.Domain == mtfs2.Domain
	destEqual := mtfs.Destination == mtfs2.Destination

	return namesEqual &&
			urlsEqual &&
			projectsEqual &&
			domainsEqual &&
			destEqual,
		nil
}

// GenerateGeneric form (map[string]interface) of the material filter
func (mtfs MaterialAttributesTfs) GenerateGeneric() (ma map[string]interface{}) {
	ma = make(map[string]interface{})
	return
}

// HasFilter in this material attribute
func (mtfs MaterialAttributesTfs) HasFilter() bool {
	return true
}

// GetFilter from material attribute
func (mtfs MaterialAttributesTfs) GetFilter() *MaterialFilter {
	return mtfs.Filter
}

// UnmarshallInterface from a JSON string to a MaterialAttributesTfs struct
func unmarshallMaterialAttributesTfs(mtfs *MaterialAttributesTfs, i map[string]interface{}) {
	for key, value := range i {
		if value == nil {
			continue
		}
		switch key {

		case "name":
			mtfs.Name = value.(string)

		case "url":
			mtfs.URL = value.(string)
		case "project_path":
			mtfs.ProjectPath = value.(string)
		case "domain":
			mtfs.Domain = value.(string)

		case "username":
			mtfs.Username = value.(string)
		case "password":
			mtfs.Password = value.(string)
		case "encrypted_password":
			mtfs.EncryptedPassword = value.(string)

		case "destination":
			mtfs.Destination = value.(string)
		case "filter":
			mtfs.Filter = unmarshallMaterialFilter(value.(map[string]interface{}))
		case "invert_filter":
			mtfs.InvertFilter = value.(bool)
		case "auto_update":
			mtfs.AutoUpdate = value.(bool)
		}
	}
}
