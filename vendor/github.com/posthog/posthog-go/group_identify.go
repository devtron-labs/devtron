package posthog

import (
	"fmt"
	"time"
)

type GroupIdentify struct {
	Type string
	Key  string

	DistinctId string
	Timestamp  time.Time
	Properties Properties
}

func (msg GroupIdentify) internal() {
	panic(unimplementedError)
}

func (msg GroupIdentify) Validate() error {
	if len(msg.Type) == 0 {
		return FieldError{
			Type:  "posthog.GroupIdentify",
			Name:  "Type",
			Value: msg.Type,
		}
	}

	if len(msg.Key) == 0 {
		return FieldError{
			Type:  "posthog.GroupIdentify",
			Name:  "Key",
			Value: msg.Key,
		}
	}

	return nil
}

type GroupIdentifyInApi struct {
	Library        string    `json:"library"`
	LibraryVersion string    `json:"library_version"`
	Timestamp      time.Time `json:"timestamp"`

	Event      string     `json:"event"`
	DistinctId string     `json:"distinct_id"`
	Properties Properties `json:"properties"`
}

func (msg GroupIdentify) APIfy() APIMessage {
	library := "posthog-go"

	myProperties := Properties{}.Set("$lib", library).Set("$lib_version", getVersion())
	myProperties.Set("$group_type", msg.Type).Set("$group_key", msg.Key).Set("$group_set", msg.Properties)

	distinctId := fmt.Sprintf("$%s_%s", msg.Type, msg.Key)

	apified := GroupIdentifyInApi{
		Event:          "$groupidentify",
		Properties:     myProperties,
		DistinctId:     distinctId,
		Timestamp:      msg.Timestamp,
		Library:        library,
		LibraryVersion: getVersion(),
	}

	return apified
}
