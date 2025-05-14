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

import (
	"errors"
	"github.com/caarlos0/env"
	repository "github.com/devtron-labs/devtron/internal/sql/repository/imageTagging"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ImageTaggingReadService interface {
	// GetImageTaggingServiceConfig returns the ImageTaggingServiceConfig
	GetImageTaggingServiceConfig() *ImageTaggingServiceConfig
	// GetTagNamesByArtifactId gets all the tag names for the given artifactId
	GetTagNamesByArtifactId(artifactId int) ([]string, error)
	// GetUniqueTagsByAppId gets all the unique tag names for the given appId
	GetUniqueTagsByAppId(appId int) ([]string, error)
}

type ImageTaggingReadServiceImpl struct {
	logger                    *zap.SugaredLogger
	imageTaggingServiceConfig *ImageTaggingServiceConfig
	imageTaggingRepo          repository.ImageTaggingRepository
}

func NewImageTaggingReadServiceImpl(
	imageTaggingRepo repository.ImageTaggingRepository,
	logger *zap.SugaredLogger) (*ImageTaggingReadServiceImpl, error) {
	imageTaggingServiceConfig := &ImageTaggingServiceConfig{}
	err := env.Parse(imageTaggingServiceConfig)
	if err != nil {
		logger.Infow("error occurred while parsing ImageTaggingServiceConfig,so setting HIDE_IMAGE_TAGGING_HARD_DELETE to default value", "err", err)
		return nil, err
	}
	return &ImageTaggingReadServiceImpl{
		logger:                    logger,
		imageTaggingServiceConfig: imageTaggingServiceConfig,
		imageTaggingRepo:          imageTaggingRepo,
	}, nil
}

func (impl *ImageTaggingReadServiceImpl) GetImageTaggingServiceConfig() *ImageTaggingServiceConfig {
	return impl.imageTaggingServiceConfig
}

func (impl *ImageTaggingReadServiceImpl) GetTagNamesByArtifactId(artifactId int) ([]string, error) {
	imageReleaseTags, err := impl.imageTaggingRepo.GetTagsByArtifactId(artifactId)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in fetching image tags using artifactId", "err", err, "artifactId", artifactId)
		return nil, err
	}
	imageLabels := make([]string, 0, len(imageReleaseTags))
	for _, imageTag := range imageReleaseTags {
		imageLabels = append(imageLabels, imageTag.TagName)
	}
	return imageLabels, nil
}

func (impl *ImageTaggingReadServiceImpl) GetUniqueTagsByAppId(appId int) ([]string, error) {
	imageTags, err := impl.imageTaggingRepo.GetTagsByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching image tags using appId", "err", err, "appId", appId)
		return nil, err
	}
	uniqueTags := make([]string, len(imageTags))
	for i, tag := range imageTags {
		uniqueTags[i] = tag.TagName
	}
	return uniqueTags, nil
}
