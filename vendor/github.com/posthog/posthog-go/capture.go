package posthog

import "time"

var _ Message = (*Capture)(nil)

// This type represents object sent in a capture call
type Capture struct {
	// This field is exported for serialization purposes and shouldn't be set by
	// the application, its value is always overwritten by the library.
	Type string
	// You don't usually need to specify this field - Posthog will generate it automatically.
	// Use it only when necessary - for example, to prevent duplicate events.
	Uuid             string
	DistinctId       string
	Event            string
	Timestamp        time.Time
	Properties       Properties
	Groups           Groups
	SendFeatureFlags bool
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
	Uuid           string    `json:"uuid"`
	Library        string    `json:"library"`
	LibraryVersion string    `json:"library_version"`
	Timestamp      time.Time `json:"timestamp"`

	DistinctId       string     `json:"distinct_id"`
	Event            string     `json:"event"`
	Properties       Properties `json:"properties"`
	SendFeatureFlags bool       `json:"send_feature_flags"`
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

	if msg.Groups != nil {
		myProperties.Set("$groups", msg.Groups)
	}

	apified := CaptureInApi{
		Type:           msg.Type,
		Uuid:           msg.Uuid,
		Library:        library,
		LibraryVersion: libraryVersion,
		Timestamp:      msg.Timestamp,
		DistinctId:     msg.DistinctId,
		Event:          msg.Event,
		Properties:     myProperties,
	}

	return apified
}
