package overlay

import (
	"context"
	"fmt"

	"github.com/speakeasy-api/openapi/internal/version"
)

// UpgradeOptions configures the upgrade behavior
type UpgradeOptions struct {
	targetVersion string
}

// Option is a functional option for configuring behavior
type Option[T any] func(*T)

// WithUpgradeTargetVersion sets the target version for the upgrade.
// If not specified, defaults to the latest supported version (1.1.0).
func WithUpgradeTargetVersion(ver string) Option[UpgradeOptions] {
	return func(opts *UpgradeOptions) {
		opts.targetVersion = ver
	}
}

// Upgrade upgrades an Overlay document from 1.0.0 to the latest version (1.1.0).
// Returns true if an upgrade was performed, false if no upgrade was needed.
//
// The upgrade process:
//   - Updates the overlay version field from 1.0.0 to 1.1.0
//   - Enables RFC 9535 JSONPath as the default implementation
//   - Clears redundant x-speakeasy-jsonpath: rfc9535 (now default in 1.1.0)
//   - All existing actions remain valid and functional
//
// Note: The upgrade is non-destructive. All 1.0.0 features continue to work in 1.1.0.
func Upgrade(_ context.Context, o *Overlay, opts ...Option[UpgradeOptions]) (bool, error) {
	if o == nil {
		return false, nil
	}

	options := UpgradeOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	if options.targetVersion == "" {
		options.targetVersion = LatestVersion
	}

	currentVersion, err := version.Parse(o.Version)
	if err != nil {
		return false, fmt.Errorf("invalid current overlay version %q: %w", o.Version, err)
	}

	targetVersion, err := version.Parse(options.targetVersion)
	if err != nil {
		return false, fmt.Errorf("invalid target overlay version %q: %w", options.targetVersion, err)
	}

	// Cannot downgrade
	if targetVersion.LessThan(*currentVersion) {
		return false, fmt.Errorf("cannot downgrade overlay version from %s to %s",
			currentVersion, targetVersion)
	}

	// Already at target version
	if targetVersion.Equal(*currentVersion) {
		return false, nil
	}

	// Apply upgrade transformations
	upgradeFrom100To110(o, currentVersion, targetVersion)

	return true, nil
}

// upgradeFrom100To110 performs the upgrade from 1.0.0 to 1.1.0
func upgradeFrom100To110(o *Overlay, currentVersion *version.Version, targetVersion *version.Version) {
	v110 := version.MustParse("1.1.0")

	if !currentVersion.LessThan(*v110) {
		return // Already at or above 1.1.0
	}

	// Calculate the max version we can upgrade to
	maxVersion := v110
	if targetVersion.LessThan(*maxVersion) {
		maxVersion = targetVersion
	}

	// The upgrade from 1.0.0 to 1.1.0:
	// 1. Update the version number
	// 2. If x-speakeasy-jsonpath was "rfc9535", it can be cleared (now default)
	// 3. No breaking changes - all 1.0.0 overlays remain valid

	// Clear explicit rfc9535 as it's now the default in 1.1.0
	if o.JSONPathVersion == JSONPathRFC9535 {
		o.JSONPathVersion = ""
	}

	o.Version = maxVersion.String()
}
