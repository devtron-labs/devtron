/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

// Package v1 implements the infra config with interface values.
package v1

type PlatformResponse struct {
	Platforms []string `json:"platforms"`
}

// internal-platforms
const (
	// RUNNER_PLATFORM is the name of the default platform; a reserved name
	RUNNER_PLATFORM = "runner"

	// Deprecated: use RUNNER_PLATFORM instead
	// CI_RUNNER_PLATFORM is earlier used as the name of the default platform
	CI_RUNNER_PLATFORM = "ci-runner"
)
