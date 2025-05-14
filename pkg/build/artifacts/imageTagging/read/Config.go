/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package read

// ImageTaggingServiceConfig holds the configuration for the image tagging service
type ImageTaggingServiceConfig struct {
	// HideImageTaggingHardDelete is a flag to hide the hard delete option in the image tagging service
	HideImageTaggingHardDelete bool `env:"HIDE_IMAGE_TAGGING_HARD_DELETE" envDefault:"false" description:"Flag to hide the hard delete option in the image tagging service"`
}

func (c *ImageTaggingServiceConfig) IsHardDeleteHidden() bool {
	if c == nil {
		// return default value of ImageTaggingServiceConfig.HideImageTaggingHardDelete
		return false
	}
	return c.HideImageTaggingHardDelete
}
