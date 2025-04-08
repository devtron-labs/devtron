# PostHog Go

Please see the main [PostHog docs](https://posthog.com/docs).

Specifically, the [Go integration](https://posthog.com/docs/integrations/go-integration) details.

# Quickstart

Install posthog to your gopath
```bash
$ go get github.com/posthog/posthog-go
```

Go 🦔!
```go
package main

import (
    "os"
    "github.com/posthog/posthog-go"
)

func main() {
    client := posthog.New(os.Getenv("POSTHOG_API_KEY")) // This value must be set to the project API key in PostHog
    // alternatively, you can do 
    // client, _ := posthog.NewWithConfig(
    //     os.Getenv("POSTHOG_API_KEY"),
    //     posthog.Config{
    //         PersonalApiKey: "your personal API key", // Set this to your personal API token you want feature flag evaluation to be more performant.  This will incur more costs, though
    //         Endpoint:       "https://us.i.posthog.com",
    //     },
    // )
    defer client.Close()

    // Capture an event
    client.Enqueue(posthog.Capture{
      DistinctId: "test-user",
      Event:      "test-snippet",
      Properties: posthog.NewProperties().
        Set("plan", "Enterprise").
        Set("friends", 42),
    })
    
    // Add context for a user
    client.Enqueue(posthog.Identify{
      DistinctId: "user:123",
      Properties: posthog.NewProperties().
        Set("email", "john@doe.com").
        Set("proUser", false),
    })
    
    // Link user contexts
    client.Enqueue(posthog.Alias{
      DistinctId: "user:123",
      Alias: "user:12345",
    })
    
    // Capture a pageview
    client.Enqueue(posthog.Capture{
      DistinctId: "test-user",
      Event:      "$pageview",
      Properties: posthog.NewProperties().
        Set("$current_url", "https://example.com"),
    })
    
    // Capture event with calculated uuid to deduplicate repeated events. 
    // The library github.com/google/uuid is used
    key := myEvent.Id + myEvent.Project
    uid := uuid.NewSHA1(uuid.NameSpaceX500, []byte(key)).String()
    client.Enqueue(posthog.Capture{
      Uuid: uid,
      DistinctId: "test-user",
      Event:      "$pageview",
      Properties: posthog.NewProperties().
        Set("$current_url", "https://example.com"),
    })

    // Check if a feature flag is enabled
    isMyFlagEnabled, err := client.IsFeatureEnabled(
            FeatureFlagPayload{
                Key:        "flag-key",
                DistinctId: "distinct_id_of_your_user",
            })

    if isMyFlagEnabled == true {
        // Do something differently for this user
    }
}

```

## Testing Locally

You can run your Go app against a local build of `posthog-go` by making the following change to your `go.mod` file for whichever your app, e.g.

```Go
module example/posthog-go-app

go 1.22.5

require github.com/posthog/posthog-go v0.0.0-20240327112532-87b23fe11103

require github.com/google/uuid v1.3.0 // indirect

replace github.com/posthog/posthog-go => /path-to-your-local/posthog-go
```

## Questions?

### [Visit the community forum.](https://posthog.com/questions)
