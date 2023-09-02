package bean

import (
	"github.com/Masterminds/semver/v3"
	"strings"
)

// ParseVersion behaves as semver.StrictNewVersion, with as sole exception
// that it allows versions with a preceding "v" (i.e. v1.2.3).
func ParseVersion(v string) (*semver.Version, error) {
	vLessV := strings.TrimPrefix(v, "v")
	if _, err := semver.StrictNewVersion(vLessV); err != nil {
		return nil, err
	}
	return semver.NewVersion(v)
}
