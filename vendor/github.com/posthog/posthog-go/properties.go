package posthog

// This type is used to represent properties in messages that support it.
// It is a free-form object so the application can set any value it sees fit but
// a few helper method are defined to make it easier to instantiate properties with
// common fields.
// Here's a quick example of how this type is meant to be used:
//
//	posthog.Page{
//		DistinctId: "0123456789",
//		Properties: posthog.NewProperties()
//			.Set("revenue", 10.0)
//			.Set("currency", "USD"),
//	}
//
type Properties map[string]interface{}

func NewProperties() Properties {
	return make(Properties, 10)
}

func (p Properties) Set(name string, value interface{}) Properties {
	p[name] = value
	return p
}
