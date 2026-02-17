/*
 * Copyright (c) 2024. Devtron Inc.
 */

package constants

// ScanEnablementType represents scan enablement filter options
type ScanEnablementType string

const (
	ScanEnabled    ScanEnablementType = "scanEnabled"    // Workflows with scanning enabled
	ScanNotEnabled ScanEnablementType = "scanNotEnabled" // Workflows with scanning disabled
)

// WorkflowSortBy represents sort field for workflow listing
type WorkflowSortBy string

const (
	WorkflowSortByWorkflowName WorkflowSortBy = "workflowName" // Sort by workflow name
	WorkflowSortByAppName      WorkflowSortBy = "appName"      // Sort by application name
	WorkflowSortByScanEnabled  WorkflowSortBy = "scanEnabled"  // Sort by scan enabled status
)

// SortOrder represents sort order
type SortOrder string

const (
	SortOrderAsc  SortOrder = "ASC"
	SortOrderDesc SortOrder = "DESC"
)
