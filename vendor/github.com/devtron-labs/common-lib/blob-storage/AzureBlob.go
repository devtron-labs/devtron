/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package blob_storage

import (
	"context"
	"fmt"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"
)

type AzureBlob struct {
}

func (impl *AzureBlob) getSharedCredentials(accountName, accountKey string) (*azblob.SharedKeyCredential, error) {
	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		log.Println("Invalid credentials with error: " + err.Error())
	}
	return credential, err
}

func (impl *AzureBlob) getTokenCredentials() (azblob.TokenCredential, error) {
	msiEndpoint, err := adal.GetMSIEndpoint()
	if err != nil {
		return nil, fmt.Errorf("failed to get the managed service identity endpoint: %v", err)
	}

	token, err := adal.NewServicePrincipalTokenFromMSI(msiEndpoint, azure.PublicCloud.ResourceIdentifiers.Storage)
	if err != nil {
		return nil, fmt.Errorf("failed to create the managed service identity token: %v", err)
	}
	err = token.Refresh()
	if err != nil {
		return nil, fmt.Errorf("failure refreshing token from MSI endpoint %w", err)
	}

	credential := azblob.NewTokenCredential(token.Token().AccessToken, impl.defaultTokenRefreshFunction(token))
	return credential, err
}

func (impl *AzureBlob) buildContainerUrl(config *AzureBlobBaseConfig, container string) (*azblob.ContainerURL, error) {
	var credential azblob.Credential
	var err error
	if len(config.AccountKey) > 0 {
		credential, err = impl.getSharedCredentials(config.AccountName, config.AccountKey)
		if err != nil {
			return nil, fmt.Errorf("failed in getting credentials: %v", err)
		}
	} else {
		credential, err = impl.getTokenCredentials()
		if err != nil {
			return nil, fmt.Errorf("failed in getting credentials: %v", err)
		}
	}
	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	// From the Azure portal, get your storage account blob service URL endpoint.
	URL, _ := url.Parse(
		fmt.Sprintf("https://%s.blob.core.windows.net/%s", config.AccountName, container))

	// Create a ContainerURL object that wraps the container URL and a request
	// pipeline to make requests.
	containerURL := azblob.NewContainerURL(*URL, p)
	return &containerURL, nil
}

func (impl *AzureBlob) DownloadBlob(context context.Context, blobName string, config *AzureBlobBaseConfig, file *os.File) (success bool, err error) {
	containerURL, err := impl.buildContainerUrl(config, config.BlobContainerName) // BlobContainerCiCache
	if err != nil {
		return false, err
	}
	res, err := containerURL.ListBlobsFlatSegment(context, azblob.Marker{}, azblob.ListBlobsSegmentOptions{
		Details: azblob.BlobListingDetails{
			Versions: false,
		},
		Prefix: blobName,
	})
	if err != nil {
		return false, err
	}
	var latestVersion string
	for _, s := range res.Segment.BlobItems {
		if s.IsCurrentVersion != nil && *s.IsCurrentVersion {
			latestVersion = *s.VersionID
			break
		}
	}
	log.Println("latest version", latestVersion)
	blobURL := containerURL.NewBlobURL(blobName).WithVersionID(latestVersion)
	err = azblob.DownloadBlobToFile(context, blobURL, 0, azblob.CountToEnd, file, azblob.DownloadFromBlobOptions{})
	return true, err
}

func (impl *AzureBlob) UploadBlob(context context.Context, blobName string, config *AzureBlobBaseConfig, inputFileName string, container string) error {
	containerURL, err := impl.buildContainerUrl(config, container)
	if err != nil {
		return err
	}
	blobURL := containerURL.NewBlockBlobURL(blobName)
	log.Println("going to upload blob url: ", blobURL, "file:", inputFileName)

	file, err := os.Open(inputFileName)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = azblob.UploadFileToBlockBlob(context, file, blobURL, azblob.UploadToBlockBlobOptions{})
	if err == nil {
		log.Println("blob uploaded successfully")
	}
	return err
}

func (impl *AzureBlob) defaultTokenRefreshFunction(spToken *adal.ServicePrincipalToken) func(credential azblob.TokenCredential) time.Duration {
	return func(credential azblob.TokenCredential) time.Duration {
		err := spToken.Refresh()
		if err != nil {
			return 0
		}
		expiresIn, err := strconv.ParseInt(string(spToken.Token().ExpiresIn), 10, 64)
		if err != nil {
			return 0
		}
		credential.SetToken(spToken.Token().AccessToken)
		return time.Duration(expiresIn-300) * time.Second
	}
}
