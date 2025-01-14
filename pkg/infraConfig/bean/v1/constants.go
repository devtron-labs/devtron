/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package v1 implements the infra config with interface values.
package v1

const GLOBAL_PROFILE_NAME = "global"

const DEFAULT_PROFILE_NAME = "default"

// QualifiedProfileMaxLength is the maximum length of an infra profile name
const QualifiedProfileMaxLength int = 50

// QualifiedDescriptionMaxLength is the maximum length of an infra profile description
const QualifiedDescriptionMaxLength int = 350

// QualifiedPlatformMaxLength is the maximum length of an infra profile platform name
const QualifiedPlatformMaxLength int = 50
