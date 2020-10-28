package gocd

import (
	"encoding/xml"
)

// ConfigTasks part of cruise-control.xml. @TODO better documentation
type ConfigTasks struct {
	Tasks []ConfigTask `xml:",any"`
}

// ConfigTask part of cruise-control.xml. @TODO better documentation
// codebeat:disable[TOO_MANY_IVARS]
type ConfigTask struct {
	// Because we need to preserve the order of tasks, and we have an array of elements with mixed types,
	// we need to use this generic xml type for tasks.
	XMLName  xml.Name        `json:",omitempty"`
	Type     string          `xml:"type,omitempty"`
	RunIf    ConfigTaskRunIf `xml:"runif"`
	Command  string          `xml:"command,attr,omitempty"  json:",omitempty"`
	Args     []string        `xml:"arg,omitempty"  json:",omitempty"`
	Pipeline string          `xml:"pipeline,attr,omitempty"  json:",omitempty"`
	Stage    string          `xml:"stage,attr,omitempty"  json:",omitempty"`
	Job      string          `xml:"job,attr,omitempty"  json:",omitempty"`
	SrcFile  string          `xml:"srcfile,attr,omitempty"  json:",omitempty"`
	SrcDir   string          `xml:"srcdir,attr,omitempty"  json:",omitempty"`
}

// codebeat:enable[TOO_MANY_IVARS]

// ConfigTaskRunIf part of cruise-control.xml. @TODO better documentation
type ConfigTaskRunIf struct {
	Status string `xml:"status,attr"`
}
