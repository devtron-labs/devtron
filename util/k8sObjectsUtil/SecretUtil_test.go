package k8sObjectsUtil

import (
	"encoding/base64"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"strings"
	"testing"
)

var (
	replacement = strings.Repeat("+", 8)
)

func TestHideSecretDataIfSecret(t *testing.T) {
	modified, err := HideValuesIfSecret(
		createSecret(map[string]string{"key1": "test", "key2": "test"}))
	require.NoError(t, err)

	assert.Equal(t, map[string]interface{}{"key1": replacement, "key2": replacement}, secretData(modified))
}

func TestHideSecretDataIfInputStringIfSecret(t *testing.T) {

	secret := createSecret(map[string]string{"key1": "test", "key2": "test"})
	outBytes, _ := runtime.Encode(unstructured.UnstructuredJSONScheme, secret)

	modified, err := HideValuesIfSecretForManifestStringInput(string(outBytes), "Secret", "")
	require.NoError(t, err)

	decoder, _ := runtime.Decode(unstructured.UnstructuredJSONScheme, []byte(modified))
	assert.Equal(t, map[string]interface{}{"key1": replacement, "key2": replacement}, secretData(decoder.(*unstructured.Unstructured)))
}

func TestHideSecretDataIfInputStringIfNotSecret(t *testing.T) {

	secret := createSecret(map[string]string{"key1": "test", "key2": "test"})
	outBytes, _ := runtime.Encode(unstructured.UnstructuredJSONScheme, secret)

	modified, err := HideValuesIfSecretForManifestStringInput(string(outBytes), "Deployment", "app")
	require.NoError(t, err)

	decoder, _ := runtime.Decode(unstructured.UnstructuredJSONScheme, []byte(modified))
	base64Encode := base64.StdEncoding.EncodeToString([]byte("test"))
	assert.Equal(t, map[string]interface{}{"key1": base64Encode, "key2": base64Encode}, secretData(decoder.(*unstructured.Unstructured)))
}

func createSecret(data map[string]string) *unstructured.Unstructured {
	secret := corev1.Secret{TypeMeta: metav1.TypeMeta{Kind: "Secret"}}
	if data != nil {
		secret.Data = make(map[string][]byte)
		for k, v := range data {
			secret.Data[k] = []byte(v)
		}
	}

	return mustToUnstructured(&secret)
}


func mustToUnstructured(obj interface{}) *unstructured.Unstructured {
	un, err := toUnstructured(obj)
	if err != nil {
		panic(err)
	}
	return un
}

func toUnstructured(obj interface{}) (*unstructured.Unstructured, error) {
	uObj, err := runtime.NewTestUnstructuredConverter(equality.Semantic).ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: uObj}, nil
}

func secretData(obj *unstructured.Unstructured) map[string]interface{} {
	data, _, _ := unstructured.NestedMap(obj.Object, "data")
	return data
}