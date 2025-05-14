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

package scanTool

import (
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/scanTool/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ScanToolMetadataService interface {
	MarkToolAsActive(toolName, version string, tx *pg.Tx) error
	MarkOtherToolsInActive(toolName string, tx *pg.Tx, version string) error
	GetActiveTool() (*repository.ScanToolMetadata, error)
	ScanToolMetadataService_ent
}

type ScanToolMetadataServiceImpl struct {
	logger                     *zap.SugaredLogger
	scanToolMetadataRepository repository.ScanToolMetadataRepository
}

func NewScanToolMetadataServiceImpl(logger *zap.SugaredLogger,
	scanToolMetadataRepository repository.ScanToolMetadataRepository) *ScanToolMetadataServiceImpl {
	return &ScanToolMetadataServiceImpl{
		logger:                     logger,
		scanToolMetadataRepository: scanToolMetadataRepository,
	}
}
func (impl *ScanToolMetadataServiceImpl) MarkToolAsActive(toolName, version string, tx *pg.Tx) error {
	return impl.scanToolMetadataRepository.MarkToolAsActive(toolName, version, tx)
}

func (impl *ScanToolMetadataServiceImpl) MarkOtherToolsInActive(toolName string, tx *pg.Tx, version string) error {
	return impl.scanToolMetadataRepository.MarkOtherToolsInActive(toolName, tx, version)
}

func (impl *ScanToolMetadataServiceImpl) GetActiveTool() (*repository.ScanToolMetadata, error) {
	return impl.scanToolMetadataRepository.FindActiveTool()
}
