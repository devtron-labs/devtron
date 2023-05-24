package bitbucket

import (
	"github.com/mitchellh/mapstructure"
)

// User is the sub struct of Client
// Reference: https://developer.atlassian.com/bitbucket/api/2/reference/resource/user
type User struct {
	c             *Client
	Uuid          string
	Username      string
	Nickname      string
	Website       string
	AccountStatus string `mapstructure:"account_status"`
	DisplayName   string `mapstructure:"display_name"`
	CreatedOn     string `mapstructure:"created_on"`
	Has2faEnabled bool   `mapstructure:"has_2fa_enabled"`
	Links         map[string]interface{}
}

// Profile is getting the user data
func (u *User) Profile() (*User, error) {
	urlStr := u.c.GetApiBaseURL() + "/user"
	response, err := u.c.execute("GET", urlStr, "")
	if err != nil {
		return nil, err
	}
	return decodeUser(response)
}

// Emails is getting user's emails
func (u *User) Emails() (interface{}, error) {
	urlStr := u.c.GetApiBaseURL() + "/user/emails"
	return u.c.execute("GET", urlStr, "")
}

func decodeUser(userResponse interface{}) (*User, error) {
	userMap := userResponse.(map[string]interface{})

	if userMap["type"] == "error" {
		return nil, DecodeError(userMap)
	}

	var user = new(User)
	err := mapstructure.Decode(userMap, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}
