package overlay

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/speakeasy-api/openapi/internal/sliceutil"
	"github.com/speakeasy-api/openapi/internal/version"
)

var (
	SupportedVersions = []*version.Version{version.MustParse("1.0.0"), version.MustParse("1.1.0")}
)

// Errors
var (
	ErrOverlayVersionInvalid                   = errors.New("overlay version is invalid")
	ErrOverlayVersionNotSupported              = fmt.Errorf("overlay version must be one of: `%s`", strings.Join(sliceutil.Map(SupportedVersions, func(v *version.Version) string { return v.String() }), ", "))
	ErrOverlayVersionMustBeDefined             = errors.New("overlay version must be defined")
	ErrOverlayInfoTitleMustBeDefined           = errors.New("overlay info title must be defined")
	ErrOverlayInfoVersionMustBeDefined         = errors.New("overlay info version must be defined")
	ErrOverlayExtendsMustBeAValidURL           = errors.New("overlay extends must be a valid URL")
	ErrOverlayMustDefineAtLeastOneAction       = errors.New("overlay must define at least one action")
	ErrOverlayActionTargetMustBeDefined        = errors.New("overlay action target must be defined")
	ErrOverlayActionRemoveAndUpdateCannotBeSet = errors.New("overlay action remove and update cannot be set")
)

type ValidationErrors []error

func (v ValidationErrors) Error() string {
	msgs := make([]string, len(v))
	for i, err := range v {
		msgs[i] = err.Error()
	}
	return strings.Join(msgs, "\n")
}

func (v ValidationErrors) Return() error {
	if len(v) > 0 {
		return v
	}
	return nil
}

func (o *Overlay) ValidateVersion() []error {
	errs := make(ValidationErrors, 0)
	overlayVersion, err := version.Parse(o.Version)
	switch {
	case err != nil || overlayVersion == nil:
		errs = append(errs, ErrOverlayVersionInvalid)
	case !overlayVersion.IsOneOf(SupportedVersions):
		errs = append(errs, ErrOverlayVersionNotSupported)
	}

	return errs
}

func (o *Overlay) Validate() error {
	errs := make(ValidationErrors, 0)

	errs = append(errs, o.ValidateVersion()...)

	if o.Info.Version == "" {
		errs = append(errs, errors.New("overlay info version must be defined"))
	}

	if o.Info.Title == "" {
		errs = append(errs, errors.New("overlay info title must be defined"))
	}

	if o.Extends != "" {
		_, err := url.Parse(o.Extends)
		if err != nil {
			errs = append(errs, errors.New("overlay extends must be a valid URL"))
		}
	}

	if len(o.Actions) == 0 {
		errs = append(errs, errors.New("overlay must define at least one action"))
	} else {
		for i, action := range o.Actions {
			if action.Target == "" {
				errs = append(errs, fmt.Errorf("overlay action at index %d target must be defined", i))
			}

			if action.Remove && !action.Update.IsZero() {
				errs = append(errs, fmt.Errorf("overlay action at index %d should not both set remove and define update", i))
			}
		}
	}

	return errs.Return()
}
