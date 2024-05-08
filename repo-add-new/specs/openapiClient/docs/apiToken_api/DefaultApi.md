# \DefaultApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**OrchestratorApiTokenGet**](DefaultApi.md#OrchestratorApiTokenGet) | **Get** /orchestrator/api-token | 
[**OrchestratorApiTokenIdDelete**](DefaultApi.md#OrchestratorApiTokenIdDelete) | **Delete** /orchestrator/api-token/{id} | 
[**OrchestratorApiTokenIdPut**](DefaultApi.md#OrchestratorApiTokenIdPut) | **Put** /orchestrator/api-token/{id} | 
[**OrchestratorApiTokenPost**](DefaultApi.md#OrchestratorApiTokenPost) | **Post** /orchestrator/api-token | 



## OrchestratorApiTokenGet

> []ApiToken OrchestratorApiTokenGet(ctx).Execute()





### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.OrchestratorApiTokenGet(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.OrchestratorApiTokenGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `OrchestratorApiTokenGet`: []ApiToken
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.OrchestratorApiTokenGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiOrchestratorApiTokenGetRequest struct via the builder pattern


### Return type

[**[]ApiToken**](ApiToken.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## OrchestratorApiTokenIdDelete

> ActionResponse OrchestratorApiTokenIdDelete(ctx, id).Execute()





### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    id := int64(789) // int64 | api-token Id

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.OrchestratorApiTokenIdDelete(context.Background(), id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.OrchestratorApiTokenIdDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `OrchestratorApiTokenIdDelete`: ActionResponse
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.OrchestratorApiTokenIdDelete`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **int64** | api-token Id | 

### Other Parameters

Other parameters are passed through a pointer to a apiOrchestratorApiTokenIdDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ActionResponse**](ActionResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## OrchestratorApiTokenIdPut

> UpdateApiTokenResponse OrchestratorApiTokenIdPut(ctx, id).UpdateApiTokenRequest(updateApiTokenRequest).Execute()





### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    id := int64(789) // int64 | api-token Id
    updateApiTokenRequest := *openapiclient.NewUpdateApiTokenRequest() // UpdateApiTokenRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.OrchestratorApiTokenIdPut(context.Background(), id).UpdateApiTokenRequest(updateApiTokenRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.OrchestratorApiTokenIdPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `OrchestratorApiTokenIdPut`: UpdateApiTokenResponse
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.OrchestratorApiTokenIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **int64** | api-token Id | 

### Other Parameters

Other parameters are passed through a pointer to a apiOrchestratorApiTokenIdPutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **updateApiTokenRequest** | [**UpdateApiTokenRequest**](UpdateApiTokenRequest.md) |  | 

### Return type

[**UpdateApiTokenResponse**](UpdateApiTokenResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## OrchestratorApiTokenPost

> CreateApiTokenResponse OrchestratorApiTokenPost(ctx).CreateApiTokenRequest(createApiTokenRequest).Execute()





### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    createApiTokenRequest := []openapiclient.CreateApiTokenRequest{*openapiclient.NewCreateApiTokenRequest()} // []CreateApiTokenRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.OrchestratorApiTokenPost(context.Background()).CreateApiTokenRequest(createApiTokenRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.OrchestratorApiTokenPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `OrchestratorApiTokenPost`: CreateApiTokenResponse
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.OrchestratorApiTokenPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiOrchestratorApiTokenPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **createApiTokenRequest** | [**[]CreateApiTokenRequest**](CreateApiTokenRequest.md) |  | 

### Return type

[**CreateApiTokenResponse**](CreateApiTokenResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

