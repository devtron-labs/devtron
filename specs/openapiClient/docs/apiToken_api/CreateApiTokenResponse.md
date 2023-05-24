# CreateApiTokenResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Success** | Pointer to **bool** | success or failure | [optional] 
**Token** | Pointer to **string** | Token of that api-token | [optional] 
**UserId** | Pointer to **int32** | User Id associated with api-token | [optional] 
**UserIdentifier** | Pointer to **string** | EmailId of that api-token user | [optional] 

## Methods

### NewCreateApiTokenResponse

`func NewCreateApiTokenResponse() *CreateApiTokenResponse`

NewCreateApiTokenResponse instantiates a new CreateApiTokenResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateApiTokenResponseWithDefaults

`func NewCreateApiTokenResponseWithDefaults() *CreateApiTokenResponse`

NewCreateApiTokenResponseWithDefaults instantiates a new CreateApiTokenResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSuccess

`func (o *CreateApiTokenResponse) GetSuccess() bool`

GetSuccess returns the Success field if non-nil, zero value otherwise.

### GetSuccessOk

`func (o *CreateApiTokenResponse) GetSuccessOk() (*bool, bool)`

GetSuccessOk returns a tuple with the Success field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSuccess

`func (o *CreateApiTokenResponse) SetSuccess(v bool)`

SetSuccess sets Success field to given value.

### HasSuccess

`func (o *CreateApiTokenResponse) HasSuccess() bool`

HasSuccess returns a boolean if a field has been set.

### GetToken

`func (o *CreateApiTokenResponse) GetToken() string`

GetToken returns the Token field if non-nil, zero value otherwise.

### GetTokenOk

`func (o *CreateApiTokenResponse) GetTokenOk() (*string, bool)`

GetTokenOk returns a tuple with the Token field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToken

`func (o *CreateApiTokenResponse) SetToken(v string)`

SetToken sets Token field to given value.

### HasToken

`func (o *CreateApiTokenResponse) HasToken() bool`

HasToken returns a boolean if a field has been set.

### GetUserId

`func (o *CreateApiTokenResponse) GetUserId() int32`

GetUserId returns the UserId field if non-nil, zero value otherwise.

### GetUserIdOk

`func (o *CreateApiTokenResponse) GetUserIdOk() (*int32, bool)`

GetUserIdOk returns a tuple with the UserId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserId

`func (o *CreateApiTokenResponse) SetUserId(v int32)`

SetUserId sets UserId field to given value.

### HasUserId

`func (o *CreateApiTokenResponse) HasUserId() bool`

HasUserId returns a boolean if a field has been set.

### GetUserIdentifier

`func (o *CreateApiTokenResponse) GetUserIdentifier() string`

GetUserIdentifier returns the UserIdentifier field if non-nil, zero value otherwise.

### GetUserIdentifierOk

`func (o *CreateApiTokenResponse) GetUserIdentifierOk() (*string, bool)`

GetUserIdentifierOk returns a tuple with the UserIdentifier field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserIdentifier

`func (o *CreateApiTokenResponse) SetUserIdentifier(v string)`

SetUserIdentifier sets UserIdentifier field to given value.

### HasUserIdentifier

`func (o *CreateApiTokenResponse) HasUserIdentifier() bool`

HasUserIdentifier returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


