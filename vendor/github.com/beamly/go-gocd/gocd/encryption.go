package gocd

import (
	"context"
)

// EncryptionService describes the HAL _link resource for the api response object for a pipelineconfig
type EncryptionService service

// CipherText sescribes the response from the api with an encrypted value.
type CipherText struct {
	EncryptedValue string    `json:"encrypted_value"`
	Links          *HALLinks `json:"_links"`
}

// Encrypt takes a plaintext value and returns a cipher text.
func (es *EncryptionService) Encrypt(ctx context.Context, plaintext string) (c *CipherText, resp *APIResponse, err error) {

	c = &CipherText{}
	_, resp, err = es.client.postAction(ctx, &APIClientRequest{
		Path:         "admin/encrypt",
		ResponseBody: c,
		RequestBody: &map[string]string{
			"value": plaintext,
		},
		APIVersion: apiV1,
	})

	return
}
