## 1.4.1

* [Full Changelog](https://github.com/PostHog/posthog-go/compare/v1.4.0...v1.4.1)

## 1.4.0

* [Full Changelog](https://github.com/PostHog/posthog-go/compare/v1.3.3...v1.4.0)

## 1.3.3

* [Full Changelog](https://github.com/PostHog/posthog-go/compare/v1.3.2...v1.3.3)

## 1.3.2

* [Full Changelog](https://github.com/PostHog/posthog-go/compare/v1.3.1...v1.3.2)

## 1.3.1

* [Full Changelog](https://github.com/PostHog/posthog-go/compare/v1.3.0...v1.3.1)

## 1.2.24

* [Full Changelog](https://github.com/PostHog/posthog-go/compare/v1.2.23...v1.2.24)

## 1.2.23

* [Full Changelog](https://github.com/PostHog/posthog-go/compare/v1.2.22...v1.2.23)

## 1.2.22

* [Full Changelog](https://github.com/PostHog/posthog-go/compare/v1.2.21...v1.2.22)

## 1.2.21

* [Full Changelog](https://github.com/PostHog/posthog-go/compare/v1.2.20...v1.2.21)

## 1.2.20

* [Full Changelog](https://github.com/PostHog/posthog-go/compare/v1.2.19...v1.2.20)

## 1.2.19

* [Full Changelog](https://github.com/PostHog/posthog-go/compare/v1.2.18...v1.2.19)

## 1.2.18

* [Full Changelog](https://github.com/PostHog/posthog-go/compare/v1.2.17...v1.2.18)

## 1.2.17

* [Full Changelog](https://github.com/PostHog/posthog-go/compare/v1.2.16...v1.2.17)

## 1.2.16

* [Full Changelog](https://github.com/PostHog/posthog-go/compare/v1.2.15...v1.2.16)

## 1.2.15

* [Full Changelog](https://github.com/PostHog/posthog-go/compare/v1.2.14...v1.2.15)

## 1.2.14

* [Full Changelog](https://github.com/PostHog/posthog-go/compare/v1.2.13...v1.2.14)

## 1.2.13

* [Full Changelog](https://github.com/PostHog/posthog-go/compare/v1.2.12...v1.2.13)

## 1.2.12

* [Full Changelog](https://github.com/PostHog/posthog-go/compare/v...v1.2.12)

## 1.2.11

* [Full Changelog](https://github.com/PostHog/posthog-go/compare/v...v1.2.11)

## 1.2.10

* [Full Changelog](https://github.com/PostHog/posthog-go/compare/v...v1.2.10)

## 1.2.9

* [Full Changelog](https://github.com/PostHog/posthog-go/compare/v...v1.2.9)

## 1.2.8

* [Full Changelog](https://github.com/PostHog/posthog-go/compare/v...v1.2.8)

## 1.2.7

* [Full Changelog](https://github.com/PostHog/posthog-go/compare/v...v1.2.7)

## 1.2.6

* [Full Changelog](https://github.com/PostHog/posthog-go/compare/v...v1.2.6)

## 1.2.5

* [Full Changelog](https://github.com/PostHog/posthog-go/compare/v...v1.2.5)

## 1.2.4

* [Full Changelog](https://github.com/PostHog/posthog-go/compare/v...v1.2.4)

## 1.2.3

* [Full Changelog](https://github.com/PostHog/posthog-go/compare/v...v1.2.3)

# Changelog

## 1.2.2 - 2024-08-08

1. Adds logging to error responses from the PostHog API so that users can see how a call failed (e.g. rate limiting) from the SDK itself.
2. Better string formatting in `poller.Errorf`

## 1.2.1 - 2024-08-07

1. The client will fall back to the `/decide` endpoint when evaluating feature flags if the user does not wish to provide a PersonalApiKey.  This fixes an issue where users were unable to use this SDK without providing a PersonalApiKey.  This fallback will make feature flag usage less performant, but will save users money by not making them pay for public API access.

## 1.2.0 - 2022-08-15

Breaking changes:

1. Minimum PostHog version requirement: 1.38
2. Local Evaluation added to IsFeatureEnabled and GetFeatureFlag. These functions now accept person and group properties arguments. The arguments will be used to locally evaluate relevant feature flags.
3. Feature flag functions take a payload called `FeatureFlagPayload` when a key is require and `FeatureFlagPayloadNoKey` when a key is not required. The payload will handle defaults for all unspecified arguments automatically.
3. Feature Flag defaults have been removed. If the flag fails for any reason, nil will be returned.
4. GetAllFlags argument added. This function returns all flags related to the id.
