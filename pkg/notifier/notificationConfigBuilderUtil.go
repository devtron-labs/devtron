package notifier

import util "github.com/devtron-labs/devtron/util/event"

type NotificationConfigRequest struct {
	Id int `json:"id"`

	TeamId    []*int `json:"teamId"`
	AppId     []*int `json:"appId"`
	EnvId     []*int `json:"envId"`
	ClusterId []*int `json:"clusterId"`

	PipelineId   *int              `json:"pipelineId"`
	PipelineType util.PipelineType `json:"pipelineType" validate:"required"`
	EventTypeIds []int             `json:"eventTypeIds" validate:"required"`
	Providers    []*Provider       `json:"providers"`
}

// GetIdsByTypeIndex
// if new criteria fields are added, add a case to this function
func (notificationConfigRequest *NotificationConfigRequest) GetIdsByTypeIndex(index typeIndex) []*int {
	switch index {
	case teams:
		return notificationConfigRequest.TeamId
	case apps:
		return notificationConfigRequest.AppId
	case envs:
		return notificationConfigRequest.EnvId
	case clusters:
		return notificationConfigRequest.ClusterId
	}
	return nil
}

// if new criteria fields are added, create a index for it here
type typeIndex int

const teams, apps, envs, clusters typeIndex = 0, 1, 2, 3

// GenerateSettingCombinations
// if new criteria is added , add a similar condition for that criteria index below
func (notificationConfigRequest *NotificationConfigRequest) GenerateSettingCombinations() []*LocalRequest {
	mask := 0
	if len(notificationConfigRequest.TeamId) > 0 {
		mask = mask | 1<<teams
	}

	if len(notificationConfigRequest.AppId) > 0 {
		mask = mask | 1<<apps
	}

	if len(notificationConfigRequest.EnvId) > 0 {
		mask = mask | 1<<envs
	}

	if len(notificationConfigRequest.ClusterId) > 0 {
		mask = mask | 1<<clusters
	}

	result := make([]*LocalRequest, 0)
	if mask == 0 {
		return append(result, &LocalRequest{PipelineId: notificationConfigRequest.PipelineId})
	}

	typeIdsArr := make([][]*int, 0)
	indices := getSetBitIndices(mask)
	for _, bitIndex := range indices {
		typeIdsArr = append(typeIdsArr, notificationConfigRequest.GetIdsByTypeIndex(typeIndex(bitIndex)))
	}
	generateCombinationSettings(typeIdsArr, LocalRequest{}, 0, &result, indices)
	return result
}

// Function to get the indices of the set bits in the mask
func getSetBitIndices(mask int) []int {
	var indices []int
	for i := 0; i < 5; i++ { // Assuming a max of 5 bits
		if mask&(1<<i) != 0 {
			indices = append(indices, i)
		}
	}
	return indices
}

// generateCombinations: add a new case if any new criteria is added
// Function to generate all combinations of arrays corresponding to the set bits
func generateCombinationSettings(arrays [][]*int, current LocalRequest, index int, result *[]*LocalRequest, indices []int) {
	// Base case: when we reach the end of the arrays, we append the combination to the result
	if index == len(arrays) {
		// Create a copy of the current result and append it
		comb := current
		*result = append(*result, &comb)
		return
	}

	// Iterate over the current array and generate combinations
	for _, value := range arrays[index] {
		valCopy := *value       // Take a copy for pointer safety
		switch indices[index] { // Set the correct index in the struct based on set bits
		case int(teams):
			current.TeamId = &valCopy
		case int(apps):
			current.AppId = &valCopy
		case int(envs):
			current.EnvId = &valCopy
		case int(clusters):
			current.ClusterId = &valCopy
		}

		generateCombinationSettings(arrays, current, index+1, result, indices)
	}
}
