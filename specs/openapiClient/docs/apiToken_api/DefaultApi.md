# \DefaultApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**OrchestratorApiTokenDelete**](DefaultApi.md#OrchestratorApiTokenDelete) | **Delete** /orchestrator/api-token | 
[**OrchestratorApiTokenGet**](DefaultApi.md#OrchestratorApiTokenGet) | **Get** /orchestrator/api-token | 
[**OrchestratorApiTokenPost**](DefaultApi.md#OrchestratorApiTokenPost) | **Post** /orchestrator/api-token | 
[**OrchestratorApiTokenPut**](DefaultApi.md#OrchestratorApiTokenPut) | **Put** /orchestrator/api-token | 



## OrchestratorApiTokenDelete

> ActionResponse OrchestratorApiTokenDelete(ctx).DeleteApiTokenRequest(deleteApiTokenRequest).Execute()





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
    deleteApiTokenRequest := *openapiclient.NewDeleteApiTokenRequest() // DeleteApiTokenRequest | 

    configuration := openapiclient.NewConfiguration()
    api_client := openapiclient.NewAPIClient(configuration)
    resp, r, err := api_client.DefaultApi.OrchestratorApiTokenDelete(context.Background()).DeleteApiTokenRequest(deleteApiTokenRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.OrchestratorApiTokenDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `OrchestratorApiTokenDelete`: ActionResponse
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.OrchestratorApiTokenDelete`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiOrchestratorApiTokenDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **deleteApiTokenRequest** | [**DeleteApiTokenRequest**](DeleteApiTokenRequest.md) |  | 

### Return type

[**ActionResponse**](ActionResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


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
    api_client := openapiclient.NewAPIClient(configuration)
    resp, r, err := api_client.DefaultApi.OrchestratorApiTokenGet(context.Background()).Execute()
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


## OrchestratorApiTokenPost

> ActionResponse OrchestratorApiTokenPost(ctx).CreateApiTokenRequest(createApiTokenRequest).Execute()





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
    api_client := openapiclient.NewAPIClient(configuration)
    resp, r, err := api_client.DefaultApi.OrchestratorApiTokenPost(context.Background()).CreateApiTokenRequest(createApiTokenRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.OrchestratorApiTokenPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `OrchestratorApiTokenPost`: ActionResponse
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

[**ActionResponse**](ActionResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## OrchestratorApiTokenPut

> ActionResponse OrchestratorApiTokenPut(ctx).UpdateApiTokenRequest(updateApiTokenRequest).Execute()





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
    updateApiTokenRequest := *openapiclient.NewUpdateApiTokenRequest() // UpdateApiTokenRequest | 

    configuration := openapiclient.NewConfiguration()
    api_client := openapiclient.NewAPIClient(configuration)
    resp, r, err := api_client.DefaultApi.OrchestratorApiTokenPut(context.Background()).UpdateApiTokenRequest(updateApiTokenRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.OrchestratorApiTokenPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `OrchestratorApiTokenPut`: ActionResponse
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.OrchestratorApiTokenPut`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiOrchestratorApiTokenPutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **updateApiTokenRequest** | [**UpdateApiTokenRequest**](UpdateApiTokenRequest.md) |  | 

### Return type

[**ActionResponse**](ActionResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

