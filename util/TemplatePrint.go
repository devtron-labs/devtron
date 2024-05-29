/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package util

import (
	"bytes"
	"text/template"
)

// Tprintf passed template string is formatted usign its operands and returns the resulting string.
// Spaces are added between operands when neither is a string.
func Tprintf(tmpl string, data interface{}) (string, error) {
	t := template.Must(template.New("tpl").Parse(tmpl))
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
