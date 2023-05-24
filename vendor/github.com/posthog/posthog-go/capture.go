package posthog

import "time"

var _ Message = (*Capture)(nil)

// This type represents object sent in a capture call
type Capture struct {
	// This field is exported for serialization purposes and shouldn't be set by
	// the application, its value is always overwritten by the library.
	Type string

	DistinctId string
	Event      string
	Timestamp  time.Time
	Properties Properties
}

func (msg Capture) internal() {
	panic(unimplementedError)
}

func (msg Capture) Validate() error {
	if len(msg.Event) == 0 {
		return FieldError{
			Type:  "posthog.Capture",
			Name:  "Event",
			Value: msg.Event,
		}
	}

	if len(msg.DistinctId) == 0 {
		return FieldError{
			Type:  "posthog.Capture",
			Name:  "DistinctId",
			Value: msg.DistinctId,
		}
	}

	return nil
}

type CaptureInApi struct {
	Type           string    `json:"type"`
	Library        string    `json:"library"`
	LibraryVersion string    `json:"library_version"`
	Timestamp      time.Time `json:"timestamp"`

	DistinctId string     `json:"distinct_id"`
	Event      string     `json:"event"`
	Properties Properties `json:"properties"`
}

func (msg Capture) APIfy() APIMessage {
	library := "posthog-go"
	libraryVersion := getVersion()

	myProperties := Properties{}.Set("$lib", library).Set("$lib_version", libraryVersion)

	if msg.Properties != nil {
		for k, v := range msg.Properties {
			myProperties[k] = v
		}
	}

	apified := CaptureInApi{
		Type:           msg.Type,
		Library:        library,
		LibraryVersion: libraryVersion,
		Timestamp:      msg.Timestamp,
		DistinctId:     msg.DistinctId,
		Event:          msg.Event,
		Properties:     myProperties,
	}

	return apified
}
