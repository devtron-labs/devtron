package bean

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"time"
)

// ------------

// Format defines the type of the VariableObject.Value
type Format string

const (
	// FormatTypeString is the string type
	FormatTypeString Format = "STRING"
	// FormatTypeNumber is the number type
	FormatTypeNumber Format = "NUMBER"
	// FormatTypeBool is the boolean type
	FormatTypeBool Format = "BOOL"
	// FormatTypeDate is the date type
	FormatTypeDate Format = "DATE"
	// FormatTypeFile is the file type
	FormatTypeFile Format = "FILE"
)

func NewFormat(format string) (Format, error) {
	return Format(format).ValuesOf(format)
}

func (d Format) ValuesOf(format string) (Format, error) {
	if format == "NUMBER" || format == "number" {
		return FormatTypeNumber, nil
	} else if format == "BOOL" || format == "bool" || format == "boolean" {
		return FormatTypeBool, nil
	} else if format == "STRING" || format == "string" {
		return FormatTypeString, nil
	} else if format == "DATE" || format == "date" {
		return FormatTypeDate, nil
	} else if format == "FILE" || format == "file" {
		return FormatTypeFile, nil
	}
	return FormatTypeString, fmt.Errorf("invalid Format: %s", format)
}

func (d Format) String() string {
	return string(d)
}

func (d Format) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Format) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	format, err := d.ValuesOf(s)
	if err != nil {
		return err
	}
	*d = format
	return nil
}

func isValidTimeStamp(candidate string, format string) bool {
	_, err := time.Parse(format, candidate)
	if err != nil {
		return false
	}
	return true
}

// isValidDateInput tries to parse the time string with multiple formats\
// and returns true if the time string is valid
func isValidDateInput(candidate string) bool {
	timeFormats := [...]string{
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		time.Kitchen,
		time.RFC3339,
		time.RFC3339Nano,
		time.DateTime,
		time.DateOnly,
		time.Stamp,
		time.StampMilli,
		time.StampMicro,
		time.StampNano,
		time.TimeOnly,
		"2006-01-02",                         // RFC 3339
		"2006-01-02 15:04",                   // RFC 3339 with minutes
		"2006-01-02 15:04:05",                // RFC 3339 with seconds
		"2006-01-02 15:04:05-07:00",          // RFC 3339 with seconds and timezone
		"2006-01-02T15Z0700",                 // ISO8601 with hour; replace Z with either + or -. use Z for UTC
		"2006-01-02T15:04Z0700",              // ISO8601 with minutes; replace Z with either + or -. use Z for UTC
		"2006-01-02T15:04:05Z0700",           // ISO8601 with seconds; replace Z with either + or -. use Z for UTC
		"2006-01-02T15:04:05.999999999Z0700", // ISO8601 with nanoseconds; replace Z with either + or -. use Z for UTC
	}
	for _, format := range timeFormats {
		if isValidTimeStamp(candidate, format) {
			return true
		}
	}
	return false
}

func (d Format) Convert(value string) (interface{}, error) {
	switch d {
	case FormatTypeString:
		return value, nil
	case FormatTypeNumber:
		return strconv.ParseFloat(value, 8)
	case FormatTypeBool:
		return strconv.ParseBool(value)
	case FormatTypeDate:
		if !isValidDateInput(value) {
			return nil, fmt.Errorf("invalid date value '%s'", value)
		}
		return value, nil
	case FormatTypeFile:
		filePath := path.Clean(value)
		fileMountDir := path.Dir(filePath)
		err := os.MkdirAll(fileMountDir, os.ModePerm)
		if err != nil {
			return nil, err
		}
		// filePath is the path to the file
		return filePath, nil
	default:
		return nil, fmt.Errorf("unsupported datatype")
	}
}

// VariableType defines the type of the VariableObject
type VariableType string

const (
	// VariableTypeValue is used to define new VariableObject value
	VariableTypeValue VariableType = "VALUE"
	// VariableTypeRefPreCi is used to refer to a VariableObject from the previous PRE-CI stage
	VariableTypeRefPreCi VariableType = "REF_PRE_CI"
	// VariableTypeRefPostCi is used to refer to a VariableObject from the previous POST-CI stage
	VariableTypeRefPostCi VariableType = "REF_POST_CI"
	// VariableTypeRefGlobal is used to refer to a VariableObject from the global scope
	VariableTypeRefGlobal VariableType = "REF_GLOBAL"
	// VariableTypeRefPlugin is used to refer to a VariableObject from the previous plugin scope
	VariableTypeRefPlugin VariableType = "REF_PLUGIN"
)

// String returns the string representation of the VariableType
func (t VariableType) String() string {
	return string(t)
}

// MarshalJSON marshals the VariableType into a JSON string
// Note: Json.Marshal will call this function internally for custom type marshalling
func (t VariableType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON unmarshal a JSON string into a VariableType
// Note: Json.Unmarshal will call this function internally for custom type unmarshalling
func (t *VariableType) UnmarshalJSON(data []byte) error {
	var variableType string
	err := json.Unmarshal(data, &variableType)
	if err != nil {
		return err
	}
	switch variableType {
	case VariableTypeValue.String(),
		VariableTypeRefPreCi.String(),
		VariableTypeRefPostCi.String(),
		VariableTypeRefGlobal.String(),
		VariableTypeRefPlugin.String():
		*t = VariableType(variableType)
		return nil
	default:
		// If the variableType is not one of the above, return an error
		// This error will be returned by the Json.Unmarshal function
		return fmt.Errorf("invalid variableType %s", data)
	}
}

// ---------------

// VariableObject defines the structure of an environment variable
//   - Name: name of the variable
//   - Format: type of the variable value.
//     Possible values are STRING, NUMBER, BOOL, DATE
//   - Value: value of the variable
//   - VariableType: defines the scope-type of the variable.
//     Possible values are VALUE, REF_PRE_CI, REF_POST_CI, REF_GLOBAL, REF_PLUGIN
//   - ReferenceVariableName: name of the variable to refer to
//   - ReferenceVariableStepIndex: index of the script step to refer to
//   - VariableStepIndexInPlugin: index of the variable in the plugin
//   - TypedValue: formatted value of the variable after type conversion.
//     This field is for internal use only (not exposed in the JSON)
type VariableObject struct {
	Name   string `json:"name"`
	Format Format `json:"format"`
	// only for input type
	Value                      string       `json:"value"`
	VariableType               VariableType `json:"variableType"`
	ReferenceVariableName      string       `json:"referenceVariableName"`
	ReferenceVariableStepIndex int          `json:"referenceVariableStepIndex"`
	VariableStepIndexInPlugin  int          `json:"variableStepIndexInPlugin"`
	FileContent                string       `json:"fileContent"` // FileContent is the base64 encoded content of the file
	TypedValue                 interface{}  `json:"-"`           // TypedValue is the formatted value of the variable after type conversion
}

// TypeCheck converts the VariableObject.Value to the VariableObject.Format type
// and stores the formatted value in the VariableObject.TypedValue field.
// If the conversion fails, it returns an error.
func (v *VariableObject) TypeCheck() error {
	typedValue, err := v.Format.Convert(v.Value)
	if err != nil {
		return err
	}
	err = v.SetFileContent(v.Value)
	if err != nil {
		return err
	}
	v.TypedValue = typedValue
	return nil
}

// SetFileContent decodes the base64 encoded file content and writes it to the file at filePath
func (v *VariableObject) SetFileContent(filePath string) error {
	if v.Format != FormatTypeFile {
		return nil
	}
	file, fileErr := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if fileErr != nil {
		return fileErr
	}
	defer file.Close()
	fileBytes, fileErr := base64.StdEncoding.DecodeString(v.FileContent)
	if fileErr != nil {
		return fileErr
	}
	_, wErr := file.Write(fileBytes)
	if wErr != nil {
		return wErr
	}
	return nil
}
