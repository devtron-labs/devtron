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

package util

import (
	"github.com/docker/distribution/reference"
	"github.com/opencontainers/go-digest"
	"log"
)

const (
	DEFAULT_TAG_VALUE = "latest"
)

type ImageMetadata struct {
	Repo   string
	Tag    string
	Digest string
}

func ExtractImageRepoAndTag(imagePath string) (imageMetadata *ImageMetadata, err error) {

	if len(imagePath) == 0 {
		return &ImageMetadata{}, nil
	}

	ref, err := reference.ParseNormalizedNamed(imagePath)
	if err != nil {
		log.Printf("error in parsing normailized ref: imagePath-%s, err - %s", imagePath, err.Error())
		return &ImageMetadata{}, err
	}

	ref = reference.TagNameOnly(ref)

	tag := getTag(ref)
	digest := getDigest(ref)
	repository := ref.Name()

	return &ImageMetadata{
		Repo:   repository,
		Tag:    tag,
		Digest: digest.String(),
	}, nil
}

func getTag(ref reference.Named) string {
	switch x := ref.(type) {
	case reference.Canonical, reference.Digested:
		return ""
	case reference.NamedTagged:
		return x.Tag()
	default:
		return ""
	}
}

func getDigest(ref reference.Named) digest.Digest {
	switch x := ref.(type) {
	case reference.Canonical:
		return x.Digest()
	case reference.Digested:
		return x.Digest()
	default:
		return digest.Digest("")
	}
}
