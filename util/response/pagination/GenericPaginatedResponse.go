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

package pagination

type SortOrder string
type SortBy string

const (
	Asc  SortOrder = "ASC"
	Desc SortOrder = "DESC"
)

const (
	AppName SortBy = "app_name"
)

type QueryParams struct {
	SortOrder SortOrder `json:"sortOrder"`
	SortBy    SortBy    `json:"sortBy"`
	Offset    int       `json:"offset"`
	Size      int       `json:"size"`
	SearchKey string    `json:"searchKey"`
}

type RepositoryRequest struct {
	Order         SortOrder
	SortBy        SortBy
	Limit, Offset int
}

type PaginatedResponse[T any] struct {
	TotalCount int `json:"totalCount"` // Total results count
	Offset     int `json:"offset"`     // Current page number
	Size       int `json:"size"`       // Current page size
	Data       []T `json:"data"`
}

// NewPaginatedResponse will initialise the PaginatedResponse; making sure that PaginatedResponse.Data will not be Null
func NewPaginatedResponse[T any]() PaginatedResponse[T] {
	return PaginatedResponse[T]{
		Data: []T{},
	}
}

// PushData will append item to the PaginatedResponse.Data
func (m *PaginatedResponse[T]) PushData(item ...T) {
	m.Data = append(m.Data, item...)
}

// UpdateTotalCount will update the TotalCount in PaginatedResponse
func (m *PaginatedResponse[_]) UpdateTotalCount(totalCount int) { // not using the type param in this method
	m.TotalCount = totalCount
}

// UpdateOffset will update the Offset in PaginatedResponse
func (m *PaginatedResponse[_]) UpdateOffset(offset int) { // not using the type param in this method
	m.Offset = offset
}

// UpdateSize will update the Size in PaginatedResponse
func (m *PaginatedResponse[_]) UpdateSize(size int) { // not using the type param in this method
	m.Size = size
}
