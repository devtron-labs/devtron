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

package git_manager

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"
)

type CiProjectDetails struct {
	GitRepository   string      `json:"gitRepository"`
	FetchSubmodules bool        `json:"fetchSubmodules"`
	MaterialName    string      `json:"materialName"`
	CheckoutPath    string      `json:"checkoutPath"`
	CommitHash      string      `json:"commitHash"`
	GitTag          string      `json:"gitTag"`
	CommitTime      time.Time   `json:"commitTime"`
	SourceType      SourceType  `json:"sourceType"`
	SourceValue     string      `json:"sourceValue"`
	Type            string      `json:"type"`
	Message         string      `json:"message"`
	Author          string      `json:"author"`
	GitOptions      GitOptions  `json:"gitOptions"`
	WebhookData     WebhookData `json:"webhookData"`
	CloningMode     string      `json:"cloningMode"`
}

func (prj *CiProjectDetails) GetCheckoutBranchName() string {
	var checkoutBranch string
	if prj.SourceType == SOURCE_TYPE_WEBHOOK {
		webhookData := prj.WebhookData
		webhookDataData := webhookData.Data

		checkoutBranch = webhookDataData[WEBHOOK_SELECTOR_TARGET_CHECKOUT_BRANCH_NAME]
		if len(checkoutBranch) == 0 {
			//webhook type is tag based
			checkoutBranch = webhookDataData[WEBHOOK_SELECTOR_TARGET_CHECKOUT_NAME]
		}
	} else {
		if len(prj.SourceValue) == 0 {
			checkoutBranch = "main"
		} else {
			checkoutBranch = prj.SourceValue
		}
	}
	if len(checkoutBranch) == 0 {
		log.Fatal("could not get target checkout from request data")
	}
	return checkoutBranch
}

type TlsPathInfo struct {
	CaCertPath  string
	TlsKeyPath  string
	TlsCertPath string
}

const (
	GIT_BASE_DIR       = "/git-base/"
	TLS_FILES_DIR      = GIT_BASE_DIR + "tls-files/"
	TLS_KEY_FILE_NAME  = "tls_key.key"
	TLS_CERT_FILE_NAME = "tls_cert.pem"
	CA_CERT_FILE_NAME  = "ca_cert.pem"
)

func getTLSKeyFileName() string {
	randomName := fmt.Sprintf("%v.key", GetRandomName())
	return randomName
}

func getCertFileName() string {
	randomName := fmt.Sprintf("%v.crt", GetRandomName())
	return randomName
}

func GetRandomName() string {
	r1 := rand.New(rand.NewSource(time.Now().UnixNano())).Int63()
	randomName := fmt.Sprintf(strconv.FormatInt(r1, 10))
	return randomName
}
