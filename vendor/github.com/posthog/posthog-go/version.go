package posthog

import "flag"

// Version of the client.
const Version = "1.4.1"

// make tests easier by using a constant version
func getVersion() string {
	if flag.Lookup("test.v") != nil {
		return "1.0.0"
	}
	return Version
}
