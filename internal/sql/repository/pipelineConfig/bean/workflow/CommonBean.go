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

package workflow

type ArtifactUploadedType string

func (r ArtifactUploadedType) String() string {
	return string(r)
}

func GetArtifactUploadedType(isUploaded bool) ArtifactUploadedType {
	if isUploaded {
		return ArtifactUploaded
	}
	return ArtifactNotUploaded
}

func IsArtifactUploaded(s ArtifactUploadedType) (isArtifactUploaded bool, isMigrationRequired bool) {
	switch s {
	case ArtifactUploaded:
		return true, false
	case ArtifactNotUploaded:
		return false, false
	default:
		return false, true
	}
}

const (
	NullArtifactUploaded ArtifactUploadedType = "NA"
	ArtifactUploaded     ArtifactUploadedType = "Uploaded"
	ArtifactNotUploaded  ArtifactUploadedType = "NotUploaded"
)
