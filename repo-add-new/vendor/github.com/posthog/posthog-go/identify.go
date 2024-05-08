package posthog

import "time"

var _ Message = (*Identify)(nil)

// This type represents object sent in an identify call
type Identify struct {
	// This field is exported for serialization purposes and shouldn't be set by
	// the application, its value is always overwritten by the library.
	Type string

	DistinctId string
	Timestamp  time.Time
	Properties Properties
}

func (msg Identify) internal() {
	panic(unimplementedError)
}

func (msg Identify) Validate() error {
	if len(msg.DistinctId) == 0 {
		return FieldError{
			Type:  "posthog.Identify",
			Name:  "DistinctId",
			Value: msg.DistinctId,
		}
	}

	return nil
}

type IdentifyInApi struct {
	Type           string    `json:"type"`
	Library        string    `json:"library"`
	LibraryVersion string    `json:"library_version"`
	Timestamp      time.Time `json:"timestamp"`

	Event      string     `json:"event"`
	DistinctId string     `json:"distinct_id"`
	Properties Properties `json:"properties"`
	Set        Properties `json:"$set"`
}

func (msg Identify) APIfy() APIMessage {
	library := "posthog-go"

	myProperties := Properties{}.Set("$lib", library).Set("$lib_version", getVersion())

	apified := IdentifyInApi{
		Type:           msg.Type,
		Event:          "$identify",
		Library:        library,
		LibraryVersion: getVersion(),
		Timestamp:      msg.Timestamp,
		DistinctId:     msg.DistinctId,

		Properties: myProperties,
		Set:        msg.Properties,
	}

	return apified
}
