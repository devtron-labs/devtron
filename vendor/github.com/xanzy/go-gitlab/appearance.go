//
// Copyright 2023, 徐晓伟 <xuxiaowei@xuxiaowei.com.cn>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package gitlab

import "net/http"

// AppearanceService handles communication with appearance of the Gitlab API.
//
// Gitlab API docs : https://docs.gitlab.com/ee/api/appearance.html
type AppearanceService struct {
	client *Client
}

// Appearance represents a GitLab appearance.
//
// Gitlab API docs : https://docs.gitlab.com/ee/api/appearance.html
type Appearance struct {
	Title                       string `json:"title"`
	Description                 string `json:"description"`
	PWAName                     string `json:"pwa_name"`
	PWAShortName                string `json:"pwa_short_name"`
	PWADescription              string `json:"pwa_description"`
	PWAIcon                     string `json:"pwa_icon"`
	Logo                        string `json:"logo"`
	HeaderLogo                  string `json:"header_logo"`
	Favicon                     string `json:"favicon"`
	NewProjectGuidelines        string `json:"new_project_guidelines"`
	ProfileImageGuidelines      string `json:"profile_image_guidelines"`
	HeaderMessage               string `json:"header_message"`
	FooterMessage               string `json:"footer_message"`
	MessageBackgroundColor      string `json:"message_background_color"`
	MessageFontColor            string `json:"message_font_color"`
	EmailHeaderAndFooterEnabled bool   `json:"email_header_and_footer_enabled"`
}

// GetAppearance gets the current appearance configuration of the GitLab instance.
//
// Gitlab API docs:
// https://docs.gitlab.com/ee/api/appearance.html#get-current-appearance-configuration
func (s *AppearanceService) GetAppearance(options ...RequestOptionFunc) (*Appearance, *Response, error) {
	req, err := s.client.NewRequest(http.MethodGet, "application/appearance", nil, options)
	if err != nil {
		return nil, nil, err
	}

	as := new(Appearance)
	resp, err := s.client.Do(req, as)
	if err != nil {
		return nil, resp, err
	}

	return as, resp, nil
}

// ChangeAppearanceOptions represents the available ChangeAppearance() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/appearance.html#change-appearance-configuration
type ChangeAppearanceOptions struct {
	Title                       *string `url:"title,omitempty" json:"title,omitempty"`
	Description                 *string `url:"description,omitempty" json:"description,omitempty"`
	PWAName                     *string `url:"pwa_name,omitempty" json:"pwa_name,omitempty"`
	PWAShortName                *string `url:"pwa_short_name,omitempty" json:"pwa_short_name,omitempty"`
	PWADescription              *string `url:"pwa_description,omitempty" json:"pwa_description,omitempty"`
	PWAIcon                     *string `url:"pwa_icon,omitempty" json:"pwa_icon,omitempty"`
	Logo                        *string `url:"logo,omitempty" json:"logo,omitempty"`
	HeaderLogo                  *string `url:"header_logo,omitempty" json:"header_logo,omitempty"`
	Favicon                     *string `url:"favicon,omitempty" json:"favicon,omitempty"`
	NewProjectGuidelines        *string `url:"new_project_guidelines,omitempty" json:"new_project_guidelines,omitempty"`
	ProfileImageGuidelines      *string `url:"profile_image_guidelines,omitempty" json:"profile_image_guidelines,omitempty"`
	HeaderMessage               *string `url:"header_message,omitempty" json:"header_message,omitempty"`
	FooterMessage               *string `url:"footer_message,omitempty" json:"footer_message,omitempty"`
	MessageBackgroundColor      *string `url:"message_background_color,omitempty" json:"message_background_color,omitempty"`
	MessageFontColor            *string `url:"message_font_color,omitempty" json:"message_font_color,omitempty"`
	EmailHeaderAndFooterEnabled *bool   `url:"email_header_and_footer_enabled,omitempty" json:"email_header_and_footer_enabled,omitempty"`
	URL                         *string `url:"url,omitempty" json:"url,omitempty"`
}

// ChangeAppearance changes the appearance configuration.
//
// Gitlab API docs:
// https://docs.gitlab.com/ee/api/appearance.html#change-appearance-configuration
func (s *AppearanceService) ChangeAppearance(opt *ChangeAppearanceOptions, options ...RequestOptionFunc) (*Appearance, *Response, error) {
	req, err := s.client.NewRequest(http.MethodPut, "application/appearance", opt, options)
	if err != nil {
		return nil, nil, err
	}

	as := new(Appearance)
	resp, err := s.client.Do(req, as)
	if err != nil {
		return nil, resp, err
	}

	return as, resp, nil
}
