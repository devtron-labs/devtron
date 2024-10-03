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
