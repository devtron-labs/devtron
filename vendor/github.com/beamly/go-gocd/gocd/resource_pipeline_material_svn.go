package gocd

import "errors"

func (mas MaterialAttributesSvn) equal(mas2i MaterialAttribute) (isEqual bool, err error) {
	var ok bool
	mas2, ok := mas2i.(MaterialAttributesSvn)
	if !ok {
		return false, errors.New("can only compare with same material type")
	}
	urlsEqual := mas.URL == mas2.URL
	destinationEqual := mas.Destination == mas2.Destination

	return urlsEqual && destinationEqual, nil
}

// GenerateGeneric form (map[string]interface) of the material filter
func (mas MaterialAttributesSvn) GenerateGeneric() (ma map[string]interface{}) {
	ma = map[string]interface{}{
		"auto_update":        mas.AutoUpdate,
		"check_externals":    mas.CheckExternals,
		"destination":        mas.Destination,
		"encrypted_password": mas.EncryptedPassword,
		"invert_filter":      mas.InvertFilter,
		"name":               mas.Name,
		"password":           mas.Password,
		"url":                mas.URL,
		"username":           mas.Username,
	}

	if f := mas.Filter.GenerateGeneric(); f != nil {
		ma["filter"] = f
	}

	return
}

// HasFilter in this material attribute
func (mas MaterialAttributesSvn) HasFilter() bool {
	return true
}

// GetFilter from material attribute
func (mas MaterialAttributesSvn) GetFilter() *MaterialFilter {
	return mas.Filter
}

// UnmarshallInterface from a JSON string to a MaterialAttributesSvn struct
func unmarshallMaterialAttributesSvn(mas *MaterialAttributesSvn, i map[string]interface{}) {
	for key, value := range i {
		if value == nil {
			continue
		}
		switch key {
		case "name":
			mas.Name = value.(string)
		case "url":
			mas.URL = value.(string)
		case "username":
			mas.Username = value.(string)
		case "password":
			mas.Password = value.(string)
		case "encrypted_password":
			mas.EncryptedPassword = value.(string)
		case "check_externals":
			mas.CheckExternals = value.(bool)
		case "destination":
			mas.Destination = value.(string)
		case "invert_filter":
			mas.InvertFilter = value.(bool)
		case "auto_update":
			mas.AutoUpdate = value.(bool)
		case "filter":
			mas.Filter = unmarshallMaterialFilter(value.(map[string]interface{}))
		}
	}
}
