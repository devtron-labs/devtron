package registry

import (
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"strings"
)

func ExtractCredentialsForRegistry(registryCredential *RegistryCredential) (string, string, error) {
	username := registryCredential.Username
	pwd := registryCredential.Password
	if (registryCredential.RegistryType == REGISTRY_TYPE_GCR || registryCredential.RegistryType == REGISTRY_TYPE_ARTIFACT_REGISTRY) && username == JSON_KEY_USERNAME {
		if strings.HasPrefix(pwd, "'") {
			pwd = pwd[1:]
		}
		if strings.HasSuffix(pwd, "'") {
			pwd = pwd[:len(pwd)-1]
		}
	}
	if registryCredential.RegistryType == DOCKER_REGISTRY_TYPE_ECR {
		accessKey, secretKey := registryCredential.AWSAccessKeyId, registryCredential.AWSSecretAccessKey
		var creds *credentials.Credentials

		if len(registryCredential.AWSAccessKeyId) == 0 || len(registryCredential.AWSSecretAccessKey) == 0 {
			sess, err := session.NewSession(&aws.Config{
				Region: &registryCredential.AWSRegion,
			})
			if err != nil {
				fmt.Printf("Error in creating AWS client", "err", err)
				return "", "", err
			}
			creds = ec2rolecreds.NewCredentials(sess)
		} else {
			creds = credentials.NewStaticCredentials(accessKey, secretKey, "")
		}
		sess, err := session.NewSession(&aws.Config{
			Region:      &registryCredential.AWSRegion,
			Credentials: creds,
		})
		if err != nil {
			fmt.Println("Error in creating AWS client session", "err", err)
			return "", "", err
		}
		svc := ecr.New(sess)
		input := &ecr.GetAuthorizationTokenInput{}
		authData, err := svc.GetAuthorizationToken(input)
		if err != nil {
			fmt.Println("Error fetching authData", "err", err)
			return "", "", err
		}
		// decode token
		token := authData.AuthorizationData[0].AuthorizationToken
		decodedToken, err := base64.StdEncoding.DecodeString(*token)
		if err != nil {
			fmt.Println("Error in decoding auth token", "err", err)
			return "", "", err
		}
		credsSlice := strings.Split(string(decodedToken), ":")
		if len(credsSlice) < 2 {
			fmt.Println("Error in decoding auth token", "err", err)
			return "", "", fmt.Errorf("error in decoding auth token for docker Registry")
		}
		username = credsSlice[0]
		pwd = credsSlice[1]
	}
	return username, pwd, nil
}
