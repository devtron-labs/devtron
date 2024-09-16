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
	"context"
	"fmt"
	"github.com/devtron-labs/common-lib/git-manager/util"
	"github.com/devtron-labs/common-lib/utils"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"path/filepath"
)

type GitOptions struct {
	UserName               string   `json:"userName"`
	Password               string   `json:"password"`
	SshPrivateKey          string   `json:"sshPrivateKey"`
	AccessToken            string   `json:"accessToken"`
	AuthMode               AuthMode `json:"authMode"`
	TlsKey                 string   `json:"tlsKey"`
	TlsCert                string   `json:"tlsCert"`
	CaCert                 string   `json:"caCert"`
	TlsVerificationEnabled bool     `json:"tlsVerificationEnabled"`
}

type WebhookData struct {
	Id              int               `json:"id"`
	EventActionType string            `json:"eventActionType"`
	Data            map[string]string `json:"data"`
}

type GitContext struct {
	context.Context // Embedding original Go context
	Auth            *BasicAuth
	WorkingDir      string
	TLSData         *TLSData
}
type BasicAuth struct {
	Username, Password string
}

type TLSData struct {
	TLSKey                 string
	TLSCertificate         string
	CACert                 string
	TlsVerificationEnabled bool
}

type AuthMode string

const (
	AUTH_MODE_USERNAME_PASSWORD AuthMode = "USERNAME_PASSWORD"
	AUTH_MODE_SSH               AuthMode = "SSH"
	AUTH_MODE_ACCESS_TOKEN      AuthMode = "ACCESS_TOKEN"
	AUTH_MODE_ANONYMOUS         AuthMode = "ANONYMOUS"
)

type SourceType string

const (
	SOURCE_TYPE_BRANCH_FIXED SourceType = "SOURCE_TYPE_BRANCH_FIXED"
	SOURCE_TYPE_WEBHOOK      SourceType = "WEBHOOK"
)

const (
	WEBHOOK_SELECTOR_TARGET_CHECKOUT_NAME        string = "target checkout"
	WEBHOOK_SELECTOR_SOURCE_CHECKOUT_NAME        string = "source checkout"
	WEBHOOK_SELECTOR_TARGET_CHECKOUT_BRANCH_NAME string = "target branch name"

	WEBHOOK_EVENT_MERGED_ACTION_TYPE     string = "merged"
	WEBHOOK_EVENT_NON_MERGED_ACTION_TYPE string = "non-merged"
)

type GitManager struct {
	GitCliManager GitCliManager
}

func NewGitManagerImpl(gitCliManager GitCliManager) *GitManager {
	return &GitManager{
		GitCliManager: gitCliManager,
	}
}

func (impl *GitManager) CloneAndCheckout(ciProjectDetails []CiProjectDetails, workingDir string) error {
	for index, prj := range ciProjectDetails {
		// git clone

		log.Println("-----> git " + prj.CloningMode + " cloning " + prj.GitRepository)

		if prj.CheckoutPath != "./" {
			if _, err := os.Stat(workingDir + prj.CheckoutPath); os.IsNotExist(err) {
				_ = os.Mkdir(workingDir+prj.CheckoutPath, os.ModeDir)
			}
		}
		var cErr error
		var auth *BasicAuth
		authMode := prj.GitOptions.AuthMode
		switch authMode {
		case AUTH_MODE_USERNAME_PASSWORD:
			auth = &BasicAuth{Password: prj.GitOptions.Password, Username: prj.GitOptions.UserName}
		case AUTH_MODE_ACCESS_TOKEN:
			auth = &BasicAuth{Password: prj.GitOptions.AccessToken, Username: prj.GitOptions.UserName}
		default:
			auth = &BasicAuth{}
		}
		tlsData := BuildTlsData(prj.GitOptions.TlsKey, prj.GitOptions.TlsCert, prj.GitOptions.CaCert, prj.GitOptions.TlsVerificationEnabled)
		gitContext := GitContext{
			Auth:       auth,
			WorkingDir: workingDir,
			TLSData:    tlsData,
		}
		// create ssh private key on disk
		if authMode == AUTH_MODE_SSH {
			cErr = util.CreateSshPrivateKeyOnDisk(index, prj.GitOptions.SshPrivateKey)
			cErr = util.CreateSshPrivateKeyOnDisk(index, prj.GitOptions.SshPrivateKey)
			if cErr != nil {
				logrus.Error("could not create ssh private key on disk ", " err ", cErr)
				return cErr
			}
		}

		_, msgMsg, cErr := impl.GitCliManager.Clone(gitContext, prj)
		if cErr != nil {
			logrus.Error("could not clone repo ", "msgMsg", msgMsg, " err ", cErr)
			return cErr
		}

		// checkout code
		if prj.SourceType == SOURCE_TYPE_BRANCH_FIXED {
			// checkout incoming commit hash or branch name
			checkoutSource := ""
			if len(prj.CommitHash) > 0 {
				checkoutSource = prj.CommitHash
			} else {
				if len(prj.SourceValue) == 0 {
					prj.SourceValue = "main"
				}
				checkoutSource = prj.SourceValue
			}
			log.Println("checkout commit in branch fix : ", checkoutSource)
			msgMsg, cErr = impl.GitCliManager.GitCheckout(gitContext, prj.CheckoutPath, checkoutSource, authMode, prj.FetchSubmodules, prj.GitRepository)
			if cErr != nil {
				logrus.Error("could not checkout hash ", "errMsg", msgMsg, "err ", cErr)
				return cErr
			}

		} else if prj.SourceType == SOURCE_TYPE_WEBHOOK {

			webhookData := prj.WebhookData
			webhookDataData := webhookData.Data

			targetCheckout := webhookDataData[WEBHOOK_SELECTOR_TARGET_CHECKOUT_NAME]
			if len(targetCheckout) == 0 {
				logrus.Error("could not get 'target checkout' from request data", "webhookData", webhookDataData)
				return fmt.Errorf("could not get 'target checkout' from request data")
			}

			log.Println("checkout commit in webhook : ", targetCheckout)

			// checkout target hash
			msgMsg, cErr = impl.GitCliManager.GitCheckout(gitContext, prj.CheckoutPath, targetCheckout, authMode, prj.FetchSubmodules, prj.GitRepository)
			if cErr != nil {
				logrus.Error("could not checkout  ", "targetCheckout ", targetCheckout, " errMsg", msgMsg, " err ", cErr)
				return cErr
			}

			// merge source if action type is merged
			if webhookData.EventActionType == WEBHOOK_EVENT_MERGED_ACTION_TYPE {
				sourceCheckout := webhookDataData[WEBHOOK_SELECTOR_SOURCE_CHECKOUT_NAME]

				// throw error if source checkout is empty
				if len(sourceCheckout) == 0 {
					logrus.Error("'source checkout' is empty", "webhookData", webhookDataData)
					return fmt.Errorf("'source checkout' is empty")
				}

				log.Println("merge commit in webhook : ", sourceCheckout)

				// merge source
				_, msgMsg, cErr = impl.GitCliManager.Merge(filepath.Join(gitContext.WorkingDir, prj.CheckoutPath), sourceCheckout)
				if cErr != nil {
					logrus.Error("could not merge ", "sourceCheckout ", sourceCheckout, " errMsg", msgMsg, " err ", cErr)
					return cErr
				}
			}
		}
	}
	return nil
}

func CreateFilesForTlsData(tlsData *TLSData, directoryPath string) (*TlsPathInfo, error) {

	if tlsData == nil {
		return nil, nil
	}
	if tlsData.TlsVerificationEnabled {
		var tlsKeyFilePath string
		var tlsCertFilePath string
		var caCertFilePath string
		var err error
		// this is to avoid concurrency issue, random number is appended at the end of file, where this file is read/created/deleted by multiple commands simultaneously.
		if tlsData.TLSKey != "" && tlsData.TLSCertificate != "" {
			tlsKeyFilePath, err = utils.CreateFolderAndFileWithContent(tlsData.TLSKey, getTLSKeyFileName(), directoryPath)
			if err != nil {
				return nil, err
			}
			tlsCertFilePath, err = utils.CreateFolderAndFileWithContent(tlsData.TLSCertificate, getCertFileName(), directoryPath)
			if err != nil {
				return nil, err
			}
		}
		if tlsData.CACert != "" {
			caCertFilePath, err = utils.CreateFolderAndFileWithContent(tlsData.CACert, getCertFileName(), directoryPath)
			if err != nil {
				return nil, err
			}
		}
		return &TlsPathInfo{caCertFilePath, tlsKeyFilePath, tlsCertFilePath}, nil
	}
	return nil, nil
}

func DeleteTlsFiles(pathInfo *TlsPathInfo) {
	if pathInfo == nil {
		return
	}
	if pathInfo.TlsKeyPath != "" {
		err := utils.DeleteAFileIfExists(pathInfo.TlsKeyPath)
		if err != nil {
			fmt.Println("error in deleting file", "tlsKeyPath", pathInfo.TlsKeyPath, "err", err)
		}
	}

	if pathInfo.TlsCertPath != "" {
		err := utils.DeleteAFileIfExists(pathInfo.TlsCertPath)
		if err != nil {
			fmt.Println("error in deleting file", "TlsCertPath", pathInfo.TlsCertPath, "err", err)
		}
	}
	if pathInfo.CaCertPath != "" {
		err := utils.DeleteAFileIfExists(pathInfo.CaCertPath)
		if err != nil {
			fmt.Println("error in deleting file", "CaCertPath", pathInfo.CaCertPath, "err", err)
		}
	}
	return
}
