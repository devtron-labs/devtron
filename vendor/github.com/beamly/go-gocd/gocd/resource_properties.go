package gocd

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"
)

type csvUnmarshal interface {
	UnmarshallCSV(raw string) error
}

type csvMarshal interface {
	MarshallCSV() (string, error)
}

// Properties describes a properties resource in the GoCD API.
type Properties struct {
	UnmarshallWithHeader bool
	IsDatum              bool
	Header               []string
	DataFrame            [][]string
}

// NewPropertiesFrame generate a new data frame for properties on a gocd job.
func NewPropertiesFrame(frame [][]string) *Properties {
	p := Properties{}
	for i, line := range frame {
		if i == 0 {
			p.Header = line
		} else {
			p.AddRow(line)
		}
	}
	return &p
}

// Get a single parameter value for a given run of the job.
func (pr Properties) Get(row int, column string) string {
	var columnIdx int
	for i, key := range pr.Header {
		if key == column {
			columnIdx = i
		}
	}
	return pr.DataFrame[row][columnIdx]
}

// AddRow to an existing properties data frame
func (pr *Properties) AddRow(r []string) {
	pr.SetRow(len(pr.DataFrame), r)
}

// SetRow in an existing data frame
func (pr *Properties) SetRow(row int, r []string) {
	for row >= len(pr.DataFrame) {
		pr.DataFrame = append(pr.DataFrame, []string{})
	}
	pr.DataFrame[row] = r
}

// MarshallCSV returns the data frame as a string
func (pr Properties) MarshallCSV() (string, error) {
	buf := new(bytes.Buffer)
	w := csv.NewWriter(buf)
	if err := w.Write(pr.Header); err != nil {
		return buf.String(), err
	}
	for _, line := range pr.DataFrame {
		if err := w.Write(line); err != nil {
			return buf.String(), err
		}
	}
	w.Flush()

	return buf.String(), nil
}

// UnmarshallCSV returns the data frame from a string
func (pr *Properties) UnmarshallCSV(raw string) error {
	r := csv.NewReader(strings.NewReader(raw))
	r.TrimLeadingSpace = true
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if pr.UnmarshallWithHeader && len(pr.Header) == 0 && len(pr.DataFrame) == 0 {
			pr.Header = record
		} else {
			pr.AddRow(record)
		}
	}
	return nil
}

// Write the data frame to a byte stream as a csv.
func (pr *Properties) Write(p []byte) (n int, err error) {
	numBytes := len(p)
	raw, err := ioutil.ReadAll(bytes.NewReader(p))
	if err != nil {
		return 0, err
	}
	pr.UnmarshallCSV(string(raw))

	return numBytes, nil
}

// MarshalJSON converts the properties structure to a list of maps
func (pr *Properties) MarshalJSON() ([]byte, error) {
	structures := make([]map[string]string, len(pr.DataFrame))

	for i, row := range pr.DataFrame {
		rowStructure := map[string]string{}
		for j, key := range pr.Header {
			value := row[j]
			rowStructure[key] = value
		}

		if pr.IsDatum {
			return json.Marshal(rowStructure)
		}

		structures[i] = rowStructure
	}

	return json.Marshal(structures)

}
