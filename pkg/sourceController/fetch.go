package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/sourceController/oci"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"k8s.io/apimachinery/pkg/util/json"
	"net/http"
	"strconv"
)

var (
	RegistryURL    string = "asia.gcr.io/google-containers"
	Insecure       bool   = true
	RepoName       string = "addon-resizer-amd64"
	ExternalCiId   int    = 3
	ImageShowCount int    = 20
)

func fetch() {
	var auth authn.Authenticator
	keychain := oci.Anonymous{}
	transport := remote.DefaultTransport.(*http.Transport).Clone()
	opts := makeRemoteOptions(context.Background(), transport, keychain, auth, true)

	url, err := parseRepositoryURLInValidFormat(RegistryURL, RepoName)
	if err != nil {
		fmt.Println("error" + err.Error())
	}
	tags, err := getAllTags(url, opts.craneOpts)
	if err != nil {
		fmt.Println("error" + err.Error())
	}
	digests := make([]string, 0, len(tags))
	digestTagMap := make(map[string]string)
	for i := 0; i < len(tags) && i < ImageShowCount; i++ {
		tag := tags[i]
		// Determine which artifact revision to pull
		tagUrl, err := getArtifactURLForTag(url, tag)
		if err != nil {
			fmt.Println("error" + err.Error())
		}
		digest, err := crane.Digest(tagUrl, opts.craneOpts...)
		if err != nil {
			fmt.Errorf("error")

		}
		digestTagMap[digest] = tag
		digests = append(digests, digest)
	}

	//err = impl.filterAlreadyPresentArtifacts(digests, digestTagMap)
	//if err != nil {
	//	fmt.Println("error"+err.Error())
	//}
	for digest, tag := range digestTagMap {
		CallExternalCIWebHook(digest, tag)
	}
	fmt.Println("success")
}

// CallingExternalCiWebhook
func CallExternalCIWebHook(digest, tag string) error {
	host := RegistryURL
	repoName := RepoName
	image := fmt.Sprintf("%s/%s:%s", host, repoName, tag)
	url := "http://172.190.239.166:30797/orchestrator/webhook/ext-ci/" + strconv.Itoa(ExternalCiId)
	payload := getPayloadForExternalCi(image, digest)
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	bearer := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTM3MzE4ODIsImp0aSI6Ijk4OThiNTNlLTU1ZTgtNDAxMS1iZjBkLWRlZjlhNGJjNWU5YSIsImlhdCI6MTY5MzY0NTQ4MiwiaXNzIjoiYXJnb2NkIiwibmJmIjoxNjkzNjQ1NDgyLCJzdWIiOiJhZG1pbiJ9.5ME76IAdLJbXhHN2u0oDcPHNmboFEFPYAUrf4rY_Ts0"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	req.Header.Set("api-token", bearer)
	req.Header.Add("Content-Type", "application/json")
	//req.Header.Add("token", bearer)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
