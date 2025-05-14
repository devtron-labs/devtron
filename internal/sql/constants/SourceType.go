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

package constants

type SourceType string

const (
	SOURCE_TYPE_BRANCH_FIXED SourceType = "SOURCE_TYPE_BRANCH_FIXED"
	SOURCE_TYPE_BRANCH_REGEX SourceType = "SOURCE_TYPE_BRANCH_REGEX"
	SOURCE_TYPE_TAG_ANY      SourceType = "SOURCE_TYPE_TAG_ANY"
	SOURCE_TYPE_WEBHOOK      SourceType = "WEBHOOK"
)
