package globalTagTests

import (
	"github.com/devtron-labs/devtron/pkg/enterprise/globalTag"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIfTagMandatoryForProjectForEmptyProjectIds(t *testing.T) {
	projectId := 1
	mandatoryProjectIdsCsv := ""
	isMandatory := globalTag.CheckIfTagIsMandatoryForProject(mandatoryProjectIdsCsv, projectId)
	assert.False(t, isMandatory)
}

func TestIfTagMandatoryForProjectForAllProjectIds(t *testing.T) {
	projectId := 1
	mandatoryProjectIdsCsv := "-1"
	isMandatory := globalTag.CheckIfTagIsMandatoryForProject(mandatoryProjectIdsCsv, projectId)
	assert.True(t, isMandatory)
}

func TestIfTagMandatoryForProjectForOtherProjectIds(t *testing.T) {
	projectId := 1
	mandatoryProjectIdsCsv := "2,3"
	isMandatory := globalTag.CheckIfTagIsMandatoryForProject(mandatoryProjectIdsCsv, projectId)
	assert.False(t, isMandatory)
}

func TestIfTagMandatoryForProjectForSameSingleProjectIds(t *testing.T) {
	projectId := 1
	mandatoryProjectIdsCsv := "1"
	isMandatory := globalTag.CheckIfTagIsMandatoryForProject(mandatoryProjectIdsCsv, projectId)
	assert.True(t, isMandatory)
}

func TestIfTagMandatoryForProjectForSameProjectIds(t *testing.T) {
	projectId := 1
	mandatoryProjectIdsCsv := "1,2"
	isMandatory := globalTag.CheckIfTagIsMandatoryForProject(mandatoryProjectIdsCsv, projectId)
	assert.True(t, isMandatory)
}

func TestCheckForValidLabelsForNilLabelsAndTags(t *testing.T) {
	err := globalTag.CheckIfValidLabels(nil, nil)
	assert.Equal(t, nil, err)
}

func TestCheckForValidLabelsForNilLabelsAndNotNilTags(t *testing.T) {
	var globalTags []*globalTag.GlobalTagDtoForProject
	err := globalTag.CheckIfValidLabels(nil, globalTags)
	assert.Equal(t, nil, err)
}

func TestCheckForValidLabelsForNotNilLabelsAndNilTags(t *testing.T) {
	labels := make(map[string]string)
	err := globalTag.CheckIfValidLabels(labels, nil)
	assert.Equal(t, nil, err)
}

func TestCheckForValidLabelsForMandatoryLabelNotPass(t *testing.T) {
	var globalTags []*globalTag.GlobalTagDtoForProject
	globalTags = append(globalTags, &globalTag.GlobalTagDtoForProject{
		Key:         "somekey",
		IsMandatory: true,
	})

	labels := make(map[string]string)
	labels["somekey2"] = "somevalue2s"

	err := globalTag.CheckIfValidLabels(labels, globalTags)
	assert.NotNil(t, err)
}

func TestCheckForValidLabelsForNoMandatoryLabels(t *testing.T) {
	var globalTags []*globalTag.GlobalTagDtoForProject
	globalTags = append(globalTags, &globalTag.GlobalTagDtoForProject{
		Key:         "somekey",
		IsMandatory: false,
	})

	labels := make(map[string]string)
	labels["somekey2"] = "somevalue2s"

	err := globalTag.CheckIfValidLabels(labels, globalTags)
	assert.Nil(t, err)
}

func TestCheckForValidLabelsForInvalidLabelKey(t *testing.T) {
	labels := make(map[string]string)
	labels["key/mid/value"] = "somevalue2s"
	err := globalTag.CheckIfValidLabels(labels, nil)
	assert.NotNil(t, err)
}

func TestCheckForValidLabelsForInvalidLabelValue(t *testing.T) {
	labels := make(map[string]string)
	labels["key"] = "value1/value2"
	err := globalTag.CheckIfValidLabels(labels, nil)
	assert.NotNil(t, err)
}

func TestCheckForValidLabels(t *testing.T) {
	labels := make(map[string]string)
	labels["key"] = "value"
	err := globalTag.CheckIfValidLabels(labels, nil)
	assert.Nil(t, err)
}
