package registry

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/sts"
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
		var sess *session.Session
		var err error

		if len(accessKey) == 0 || len(secretKey) == 0 {
			// Case 1: IAM role — use default credential chain (IRSA, instance profile, task role, env vars)
			sess, err = session.NewSession(&aws.Config{
				Region: &registryCredential.AWSRegion,
			})
		} else {
			// Case 2: Static credentials
			creds := credentials.NewStaticCredentials(accessKey, secretKey, "")
			sess, err = session.NewSession(&aws.Config{
				Region:      &registryCredential.AWSRegion,
				Credentials: creds,
			})
		}
		if err != nil {
			fmt.Println("Error in creating AWS client session", "err", err)
			return "", "", err
		}

		// Case 3: AssumeRole (cross-account) — layered on top of Case 1 or 2
		if len(registryCredential.AssumeRoleArn) > 0 {
			stsClient := sts.New(sess)
			assumeOutput, err := stsClient.AssumeRole(&sts.AssumeRoleInput{
				RoleArn:         aws.String(registryCredential.AssumeRoleArn),
				RoleSessionName: aws.String("devtron-ecr-cross-account"),
			})
			if err != nil {
				fmt.Printf("Error in assuming role %s: %v", registryCredential.AssumeRoleArn, err)
				return "", "", err
			}
			assumedCreds := credentials.NewStaticCredentials(
				*assumeOutput.Credentials.AccessKeyId,
				*assumeOutput.Credentials.SecretAccessKey,
				*assumeOutput.Credentials.SessionToken,
			)
			sess, err = session.NewSession(&aws.Config{
				Region:      &registryCredential.AWSRegion,
				Credentials: assumedCreds,
			})
			if err != nil {
				fmt.Println("Error in creating AWS session with assumed role credentials", "err", err)
				return "", "", err
			}
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
