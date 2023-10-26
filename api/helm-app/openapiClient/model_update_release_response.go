/*
Devtron Labs

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: 1.0.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi

import (
	"encoding/json"
)

// UpdateReleaseResponse struct for UpdateReleaseResponse
type UpdateReleaseResponse struct {
	// success or failure
	Success *bool `json:"success,omitempty"`
	PerformedHelmSyncInstall *bool `json:"performedHelmSyncInstall,omitempty"`
}

// NewUpdateReleaseResponse instantiates a new UpdateReleaseResponse object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewUpdateReleaseResponse() *UpdateReleaseResponse {
	this := UpdateReleaseResponse{}
	return &this
}

// NewUpdateReleaseResponseWithDefaults instantiates a new UpdateReleaseResponse object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewUpdateReleaseResponseWithDefaults() *UpdateReleaseResponse {
	this := UpdateReleaseResponse{}
	return &this
}

// GetSuccess returns the Success field value if set, zero value otherwise.
func (o *UpdateReleaseResponse) GetSuccess() bool {
	if o == nil || o.Success == nil {
		var ret bool
		return ret
	}
	return *o.Success
}

// GetSuccessOk returns a tuple with the Success field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UpdateReleaseResponse) GetSuccessOk() (*bool, bool) {
	if o == nil || o.Success == nil {
		return nil, false
	}
	return o.Success, true
}

// HasSuccess returns a boolean if a field has been set.
func (o *UpdateReleaseResponse) HasSuccess() bool {
	if o != nil && o.Success != nil {
		return true
	}

	return false
}

// SetSuccess gets a reference to the given bool and assigns it to the Success field.
func (o *UpdateReleaseResponse) SetSuccess(v bool) {
	o.Success = &v
}

func (o UpdateReleaseResponse) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.Success != nil {
		toSerialize["success"] = o.Success
	}
	return json.Marshal(toSerialize)
}

type NullableUpdateReleaseResponse struct {
	value *UpdateReleaseResponse
	isSet bool
}

func (v NullableUpdateReleaseResponse) Get() *UpdateReleaseResponse {
	return v.value
}

func (v *NullableUpdateReleaseResponse) Set(val *UpdateReleaseResponse) {
	v.value = val
	v.isSet = true
}

func (v NullableUpdateReleaseResponse) IsSet() bool {
	return v.isSet
}

func (v *NullableUpdateReleaseResponse) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableUpdateReleaseResponse(val *UpdateReleaseResponse) *NullableUpdateReleaseResponse {
	return &NullableUpdateReleaseResponse{value: val, isSet: true}
}

func (v NullableUpdateReleaseResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableUpdateReleaseResponse) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


