package cluster

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/errors"
	testclient "k8s.io/client-go/kubernetes/fake"
	"net/url"
	"testing"
)

func TestClusterService(t *testing.T) {
	t.Run("CheckIfConfigIsValid", func(t *testing.T) {
		clientSet := testclient.NewSimpleClientset()
		response, err := clientSet.Discovery().RESTClient().Get().AbsPath("/livez").DoRaw(context.Background())
		var responseErr error
		if err != nil {
			if _, ok := err.(*url.Error); ok {
				responseErr = fmt.Errorf("Incorrect server url : %v", err)
			} else if statusError, ok := err.(*errors.StatusError); ok {
				if statusError != nil {
					responseErr = fmt.Errorf("%s : %s", statusError.ErrStatus.Reason, statusError.ErrStatus.Message)
				} else {
					responseErr = fmt.Errorf("Validation failed : %v", err)
				}
			} else {
				responseErr = fmt.Errorf("Validation failed : %v", err)
			}
		} else if err == nil && string(response) != "ok" {
			responseErr = fmt.Errorf("Validation failed with response : %s", string(response))
		}
		assert.NotNil(t, responseErr)
	})
}
