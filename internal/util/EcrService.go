/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package util

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/juju/errors"
)

//FIXME: this code is temp
func CreateEcrRepo(repoName string, RegistryURL string, reg string, accessKey string, secretKey string) error {
	region := reg
	credentials := credentials.NewStaticCredentials(accessKey, secretKey, "")
	svc := ecr.New(session.New(&aws.Config{
		Region:      &region,
		Credentials: credentials,
	}))
	input := &ecr.CreateRepositoryInput{
		RepositoryName: aws.String(repoName),
	}
	result, err := svc.CreateRepository(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ecr.ErrCodeServerException:
				fmt.Println(ecr.ErrCodeServerException, aerr.Error())
			case ecr.ErrCodeInvalidParameterException:
				fmt.Println(ecr.ErrCodeInvalidParameterException, aerr.Error())
			case ecr.ErrCodeInvalidTagParameterException:
				fmt.Println(ecr.ErrCodeInvalidTagParameterException, aerr.Error())
			case ecr.ErrCodeTooManyTagsException:
				fmt.Println(ecr.ErrCodeTooManyTagsException, aerr.Error())
			case ecr.ErrCodeRepositoryAlreadyExistsException:
				fmt.Println(ecr.ErrCodeRepositoryAlreadyExistsException, aerr.Error())
				return errors.NewAlreadyExists(aerr, "repoName")
			case ecr.ErrCodeLimitExceededException:
				fmt.Println(ecr.ErrCodeLimitExceededException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return err
	}

	fmt.Println(result)
	return err
}
