# PostHog Go

[![Go Reference](https://pkg.go.dev/badge/github.com/posthog/posthog-go.svg)](https://pkg.go.dev/github.com/posthog/posthog-go)

Please see the main [PostHog docs](https://posthog.com/docs).

Specifically, the [Go integration](https://posthog.com/docs/integrations/go-integration) details.

## Quickstart

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

## Development

Make sure you have Go installed (macOS: `brew install go`, Linx / Windows: https://go.dev/doc/install).

To build the project:

```bash
# Install dependencies
make dependencies

# Run tests and build
make build

# Just run tests
make test
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

## Examples

Check out the [examples](examples/README.md) for more detailed examples of how to use the PostHog Go client.

## Running the examples

The examples demonstrate different features of the PostHog Go client. To run all examples:

```bash
# Set your PostHog API keys and endpoint (optional)
export POSTHOG_PROJECT_API_KEY="your-project-api-key"
export POSTHOG_PERSONAL_API_KEY="your-personal-api-key"
export POSTHOG_ENDPOINT="https://app.posthog.com"  # Optional, defaults to http://localhost:8000

# Run all examples
go run examples/*.go
```

This will run:

- Feature flags example
- Capture events example
- Capture events with feature flag options example

### Prerequisites

Before running the examples, you'll need to:

1. Have a PostHog instance running (default: http://localhost:8000)
   - You can modify the endpoint by setting the `POSTHOG_ENDPOINT` environment variable
   - If not set, it defaults to "http://localhost:8000"

2. Set up the following feature flags in your PostHog instance:
   - `multivariate-test` (a multivariate flag)
   - `simple-test` (a simple boolean flag)
   - `multivariate-simple-test` (a multivariate flag)
   - `my_secret_flag_value` (a remote config flag with string payload)
   - `my_secret_flag_json_object_value` (a remote config flag with JSON object payload)
   - `my_secret_flag_json_array_value` (a remote config flag with JSON array payload)

3. Set your PostHog API keys as environment variables:
   - `POSTHOG_PROJECT_API_KEY`: Your project API key (starts with `phc_...`)
   - `POSTHOG_PERSONAL_API_KEY`: Your personal API key (starts with `phx_...`)

## Releasing

To release a new version of the PostHog Go client, follow these steps:

1. Update the version in the `version.go` file
2. Update the changelog in `CHANGELOG.md`
3. Once your changes are merged into main, create a new tag with the new version

```bash
git tag v1.4.7
git push --tags
```

4. [create a new release on GitHub](https://github.com/PostHog/posthog-go/releases/new).

Releases are installed directly from GitHub.

## Questions?

### [Visit the community forum.](https://posthog.com/questions)
