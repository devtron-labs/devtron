package cmdUtil

import "fmt"

// SanitizeCliParam is used where we are directly injecting the user defined params to any CLI commands. This prevents any script injection to the running env
func SanitizeCliParam(param string) string {
	return fmt.Sprintf("%q", param)
}
