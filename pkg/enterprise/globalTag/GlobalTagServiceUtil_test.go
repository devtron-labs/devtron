package globalTag

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIfTagMandatoryForProjectForEmptyProjectIds(t *testing.T) {
	projectId := 1
	mandatoryProjectIdsCsv := ""
	isMandatory := CheckIfTagIsMandatoryForProject(mandatoryProjectIdsCsv, projectId)
	assert.False(t, isMandatory)
}

func TestIfTagMandatoryForProjectForAllProjectIds(t *testing.T) {
	projectId := 1
	mandatoryProjectIdsCsv := "-1"
	isMandatory := CheckIfTagIsMandatoryForProject(mandatoryProjectIdsCsv, projectId)
	assert.True(t, isMandatory)
}

func TestIfTagMandatoryForProjectForOtherProjectIds(t *testing.T) {
	projectId := 1
	mandatoryProjectIdsCsv := "2,3"
	isMandatory := CheckIfTagIsMandatoryForProject(mandatoryProjectIdsCsv, projectId)
	assert.False(t, isMandatory)
}

func TestIfTagMandatoryForProjectForSameSingleProjectIds(t *testing.T) {
	projectId := 1
	mandatoryProjectIdsCsv := "1"
	isMandatory := CheckIfTagIsMandatoryForProject(mandatoryProjectIdsCsv, projectId)
	assert.True(t, isMandatory)
}

func TestIfTagMandatoryForProjectForSameProjectIds(t *testing.T) {
	projectId := 1
	mandatoryProjectIdsCsv := "1,2"
	isMandatory := CheckIfTagIsMandatoryForProject(mandatoryProjectIdsCsv, projectId)
	assert.True(t, isMandatory)
}

func TestCheckForValidLabelsForNilLabelsAndTags(t *testing.T) {
	err := CheckIfValidLabels(nil, nil)
	assert.Equal(t, nil, err)
}

func TestCheckForValidLabelsForNilLabelsAndNotNilTags(t *testing.T) {
	var globalTags []*GlobalTagDtoForProject
	err := CheckIfValidLabels(nil, globalTags)
	assert.Equal(t, nil, err)
}

func TestCheckForValidLabelsForNotNilLabelsAndNilTags(t *testing.T) {
	labels := make(map[string]string)
	err := CheckIfValidLabels(labels, nil)
	assert.Equal(t, nil, err)
}

func TestCheckForValidLabelsForMandatoryLabelNotPass(t *testing.T) {
	var globalTags []*GlobalTagDtoForProject
	globalTags = append(globalTags, &GlobalTagDtoForProject{
		Key:         "somekey",
		IsMandatory: true,
	})

	labels := make(map[string]string)
	labels["somekey2"] = "somevalue2s"

	err := CheckIfValidLabels(labels, globalTags)
	assert.NotNil(t, err)
}

func TestCheckForValidLabelsForNoMandatoryLabels(t *testing.T) {
	var globalTags []*GlobalTagDtoForProject
	globalTags = append(globalTags, &GlobalTagDtoForProject{
		Key:         "somekey",
		IsMandatory: false,
	})

	labels := make(map[string]string)
	labels["somekey2"] = "somevalue2s"

	err := CheckIfValidLabels(labels, globalTags)
	assert.Nil(t, err)
}

func TestCheckForValidLabelsForInvalidLabelKey(t *testing.T) {
	labels := make(map[string]string)
	labels["key/mid/value"] = "somevalue2s"
	err := CheckIfValidLabels(labels, nil)
	assert.NotNil(t, err)
}

func TestCheckForValidLabelsForInvalidLabelValue(t *testing.T) {
	labels := make(map[string]string)
	labels["key"] = "value1/value2"
	err := CheckIfValidLabels(labels, nil)
	assert.NotNil(t, err)
}

func TestCheckForValidLabels(t *testing.T) {
	labels := make(map[string]string)
	labels["key"] = "value"
	err := CheckIfValidLabels(labels, nil)
	assert.Nil(t, err)
}
