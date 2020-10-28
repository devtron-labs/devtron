package gocd

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Equal is true if the two materials are logically equivalent. Not neccesarily literally equal.
func (m Material) Equal(a *Material) (isEqual bool, err error) {
	if m.Type != a.Type {
		return
	}

	return m.Attributes.equal(a.Attributes)
}

// UnmarshalJSON string into a Material struct
func (m *Material) UnmarshalJSON(b []byte) (err error) {
	raw := map[string]interface{}{}
	if err = json.Unmarshal(b, &raw); err == nil {
		err = m.Ingest(raw)
	}

	return
}

// Ingest an abstract structure
func (m *Material) Ingest(payload map[string]interface{}) (err error) {

	m.IngestType(payload)

	for key, value := range payload {
		if value == nil {
			continue
		}
		switch key {
		case "attributes":
			err = m.IngestAttributeGenerics(value)
		case "fingerprint":
			m.Fingerprint = value.(string)
		case "description":
			m.Description = value.(string)
		case "type":
			continue
		default:
			err = fmt.Errorf("unexpected key: '%s'", key)
		}
		if err != nil {
			break
		}
	}
	return
}

// IngestType of Material if it is provided
func (m *Material) IngestType(payload map[string]interface{}) {
	if mType, hasMType := payload["type"]; hasMType {
		m.Type = mType.(string)
		m.IngestAttributes(map[string]interface{}{})
	}
}

// IngestAttributeGenerics to Material and perform some error checking
func (m *Material) IngestAttributeGenerics(i interface{}) (err error) {
	if v1, ok1 := i.(map[string]interface{}); ok1 {
		err = m.IngestAttributes(v1)
	}
	return
}

// IngestAttributes to Material from an abstract structure
func (m *Material) IngestAttributes(rawAttributes map[string]interface{}) (err error) {
	switch strings.ToLower(m.Type) {
	case "git":
		mag := &MaterialAttributesGit{}
		unmarshallMaterialAttributesGit(mag, rawAttributes)
		m.Attributes = mag
	case "svn":
		mas := &MaterialAttributesSvn{}
		unmarshallMaterialAttributesSvn(mas, rawAttributes)
		m.Attributes = mas
	case "hg":
		mah := &MaterialAttributesHg{}
		unmarshallMaterialAttributesHg(mah, rawAttributes)
		m.Attributes = mah
	case "p4":
		map4 := &MaterialAttributesP4{}
		unmarshallMaterialAttributesP4(map4, rawAttributes)
		m.Attributes = map4
	case "tfs":
		mat := &MaterialAttributesTfs{}
		unmarshallMaterialAttributesTfs(mat, rawAttributes)
		m.Attributes = mat
	case "dependency":
		mad := &MaterialAttributesDependency{}
		unmarshallMaterialAttributesDependency(mad, rawAttributes)
		m.Attributes = mad
	case "package":
		mapp := &MaterialAttributesPackage{}
		unmarshallMaterialAttributesPackage(mapp, rawAttributes)
		m.Attributes = mapp
	case "plugin":
		mapl := &MaterialAttributesPlugin{}
		unmarshallMaterialAttributesPlugin(mapl, rawAttributes)
		m.Attributes = mapl
	default:
		err = fmt.Errorf("Unexpected Material type: '%s'", m.Type)
	}

	return
}

// GenerateGeneric form (map[string]interface) of the material filter
func (mf *MaterialFilter) GenerateGeneric() (g map[string]interface{}) {
	if mf != nil {
		ignores := []interface{}{}
		for _, ig := range mf.Ignore {
			ignores = append(ignores, ig)
		}
		g = map[string]interface{}{
			"ignore": ignores,
		}
	}
	return
}

func unmarshallMaterialFilter(i map[string]interface{}) (m *MaterialFilter) {
	m = &MaterialFilter{}
	if ignoreI, ok1 := i["ignore"]; ok1 {
		if ignoreIs, ok2 := ignoreI.([]interface{}); ok2 {
			m.Ignore = decodeConfigStringList(ignoreIs)
		} else if ignores, ok3 := ignoreI.([]string); ok3 {
			m.Ignore = ignores
		}
	}
	return
}

// Give an abstract list of strings cast as []interface{}, convert them back to []string{}.
func decodeConfigStringList(lI []interface{}) (ret []string) {

	if len(lI) == 1 {
		return []string{lI[0].(string)}
	}
	ret = make([]string, len(lI))
	for i, vI := range lI {
		ret[i] = vI.(string)
	}
	return
}
