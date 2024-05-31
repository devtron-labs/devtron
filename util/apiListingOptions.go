/*
 * Copyright (c) 2024. Devtron Inc.
 */

package util

type ListingFilterOptions struct {
	// list filter data
	Limit        int
	Offset       int
	SearchString string
	Order        string
	SortBy       string
}

func (opts ListingFilterOptions) GetSearchStringRegex() string {
	return "%" + opts.SearchString + "%"
}
