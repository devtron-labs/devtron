package helper

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"golang.org/x/mod/semver"
	"net/http"
	"strings"
)

func CheckIfReleaseVersionIsValid(releaseVersionForValidation string) error {
	if !strings.HasPrefix(releaseVersionForValidation, "v") { //checking this because FE only sends version
		releaseVersionForValidation = fmt.Sprintf("v%s", releaseVersionForValidation)
	}
	if !semver.IsValid(releaseVersionForValidation) || len(strings.Split(releaseVersionForValidation, ".")) != 3 {
		return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.ReleaseVersionNotValid, bean.ReleaseVersionNotValid)
	}
	return nil
}
