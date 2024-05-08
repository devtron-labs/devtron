# UpdateApiTokenResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Success** | Pointer to **bool** | success or failure | [optional] 
**Token** | Pointer to **string** | Token of that api-token | [optional] 

## Methods

### NewUpdateApiTokenResponse

`func NewUpdateApiTokenResponse() *UpdateApiTokenResponse`

NewUpdateApiTokenResponse instantiates a new UpdateApiTokenResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateApiTokenResponseWithDefaults

`func NewUpdateApiTokenResponseWithDefaults() *UpdateApiTokenResponse`

NewUpdateApiTokenResponseWithDefaults instantiates a new UpdateApiTokenResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSuccess

`func (o *UpdateApiTokenResponse) GetSuccess() bool`

GetSuccess returns the Success field if non-nil, zero value otherwise.

### GetSuccessOk

`func (o *UpdateApiTokenResponse) GetSuccessOk() (*bool, bool)`

GetSuccessOk returns a tuple with the Success field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSuccess

`func (o *UpdateApiTokenResponse) SetSuccess(v bool)`

SetSuccess sets Success field to given value.

### HasSuccess

`func (o *UpdateApiTokenResponse) HasSuccess() bool`

HasSuccess returns a boolean if a field has been set.

### GetToken

`func (o *UpdateApiTokenResponse) GetToken() string`

GetToken returns the Token field if non-nil, zero value otherwise.

### GetTokenOk

`func (o *UpdateApiTokenResponse) GetTokenOk() (*string, bool)`

GetTokenOk returns a tuple with the Token field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToken

`func (o *UpdateApiTokenResponse) SetToken(v string)`

SetToken sets Token field to given value.

### HasToken

`func (o *UpdateApiTokenResponse) HasToken() bool`

HasToken returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


