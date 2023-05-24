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

package git

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

type WebhookSecretValidator interface {
	ValidateSecret(r *http.Request, secretInUrl string, requestBodyBytes []byte, gitHost *pipeline.GitHostRequest) bool
}

type WebhookSecretValidatorImpl struct {
	logger *zap.SugaredLogger
}

func NewWebhookSecretValidatorImpl(Logger *zap.SugaredLogger) *WebhookSecretValidatorImpl {
	return &WebhookSecretValidatorImpl{
		logger: Logger,
	}
}

const (
	SECRET_VALIDATOR_SHA1       string = "SHA-1"
	SECRET_VALIDATOR_URL_APPEND string = "URL_APPEND"
	SECRET_VALIDATOR_PLAIN_TEXT string = "PLAIN_TEXT"
)

// Validate secret for some predefined algorithms : SHA1, URL_APPEND, PLAIN_TEXT
// URL_APPEND : Secret will come in URL (last path param of URL)
// PLAIN_TEXT : Plain text value in request header
// SHA1 : SHA1 encrypted text in request header
func (impl *WebhookSecretValidatorImpl) ValidateSecret(r *http.Request, secretInUrl string, requestBodyBytes []byte, gitHost *pipeline.GitHostRequest) bool {

	secretValidator := gitHost.SecretValidator
	impl.logger.Debug("Validating signature for secret validator : ", secretValidator)

	switch secretValidator {

	case SECRET_VALIDATOR_SHA1:

		gotHash := strings.SplitN(r.Header.Get(gitHost.SecretHeader), "=", 2)
		if gotHash[0] != "sha1" {
			return false
		}
		hash := hmac.New(sha1.New, []byte(gitHost.WebhookSecret))
		if _, err := hash.Write(requestBodyBytes); err != nil {
			return false
		}
		expectedHash := hex.EncodeToString(hash.Sum(nil))
		return gotHash[1] == expectedHash

	case SECRET_VALIDATOR_URL_APPEND:
		secretFromUrlFromDb := gitHost.WebhookUrl[strings.LastIndex(gitHost.WebhookUrl, "/")+1:]
		return secretInUrl == secretFromUrlFromDb

	case SECRET_VALIDATOR_PLAIN_TEXT:
		secretHeaderValue := r.Header.Get(gitHost.SecretHeader)
		return secretHeaderValue == gitHost.WebhookSecret

	default:
		impl.logger.Errorw("unsupported SecretValidator ", "SecretValidator", gitHost.SecretValidator)
	}

	return false
}
