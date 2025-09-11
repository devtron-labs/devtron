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

package providers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/devtron-labs/common-lib/cloud-provider-identifier/bean"
	"go.uber.org/zap"
)

type instanceIdentityResponse struct {
	ImageID          string `json:"imageId"`
	InstanceID       string `json:"instanceId"`
	AvailabilityZone string `json:"availabilityZone"`
	Region           string `json:"region"`
	InstanceType     string `json:"instanceType"`
}

type IdentifyAmazon struct {
	Logger *zap.SugaredLogger
}

// AWS environment variables commonly available in EKS environments
var awsEnvironmentVariables = []string{
	"AWS_REGION",
	"AWS_DEFAULT_REGION",
	"AWS_ROLE_ARN",
	"AWS_WEB_IDENTITY_TOKEN_FILE",
	"AWS_STS_REGIONAL_ENDPOINTS",
}

// AWS service account token paths for EKS
var awsServiceAccountPaths = []string{
	"/var/run/secrets/eks.amazonaws.com/serviceaccount/token",
	"/var/run/secrets/kubernetes.io/serviceaccount/token",
}

func (impl *IdentifyAmazon) Identify() (string, error) {
	// Try multiple detection methods in order of reliability for EKS environments

	// 1. Check AWS environment variables (most reliable for EKS)
	if impl.checkAWSEnvironmentVariables() {
		impl.Logger.Infow("AWS detected via environment variables")
		return bean.Amazon, nil
	}

	// 2. Check for AWS service account tokens (EKS-specific)
	if impl.checkAWSServiceAccountTokens() {
		impl.Logger.Infow("AWS detected via service account tokens")
		return bean.Amazon, nil
	}

	// 3. Check /proc/version for AWS-specific information
	if impl.checkProcVersion() {
		impl.Logger.Infow("AWS detected via /proc/version")
		return bean.Amazon, nil
	}

	// 4. Check DMI system files (backward compatibility for EC2)
	if impl.checkDMISystemFiles() {
		impl.Logger.Infow("AWS detected via DMI system files")
		return bean.Amazon, nil
	}

	impl.Logger.Debugw("AWS not detected via file-based methods")
	return bean.Unknown, nil
}

// checkAWSEnvironmentVariables checks for AWS-specific environment variables
func (impl *IdentifyAmazon) checkAWSEnvironmentVariables() bool {
	for _, envVar := range awsEnvironmentVariables {
		if value := os.Getenv(envVar); value != "" {
			impl.Logger.Debugw("Found AWS environment variable", "variable", envVar, "value", value)
			return true
		}
	}
	return false
}

// checkAWSServiceAccountTokens checks for AWS service account tokens (EKS-specific)
func (impl *IdentifyAmazon) checkAWSServiceAccountTokens() bool {
	for _, tokenPath := range awsServiceAccountPaths {
		if _, err := os.Stat(tokenPath); err == nil {
			impl.Logger.Debugw("Found AWS service account token", "path", tokenPath)
			return true
		}
	}
	return false
}

// checkProcVersion checks /proc/version for AWS-specific information
func (impl *IdentifyAmazon) checkProcVersion() bool {
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		impl.Logger.Debugw("Could not read /proc/version", "error", err)
		return false
	}

	content := strings.ToLower(string(data))
	awsIndicators := []string{"aws", "amazon", "ec2", "xen"}

	for _, indicator := range awsIndicators {
		if strings.Contains(content, indicator) {
			impl.Logger.Debugw("Found AWS indicator in /proc/version", "indicator", indicator)
			return true
		}
	}
	return false
}

// checkDMISystemFiles checks DMI system files (backward compatibility)
func (impl *IdentifyAmazon) checkDMISystemFiles() bool {
	data, err := os.ReadFile(bean.AmazonSysFile)
	if err != nil {
		impl.Logger.Debugw("Could not read DMI system file", "file", bean.AmazonSysFile, "error", err)
		return false
	}

	if strings.Contains(string(data), bean.AmazonIdentifierString) {
		impl.Logger.Debugw("Found AWS identifier in DMI system file", "identifier", bean.AmazonIdentifierString)
		return true
	}
	return false
}

func (impl *IdentifyAmazon) IdentifyViaMetadataServer(detected chan<- string) {
	// Create HTTP client with timeout for metadata service
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Try to get IMDSv2 token first
	token, err := impl.getIMDSv2Token(client)
	if err != nil {
		impl.Logger.Debugw("Failed to get IMDSv2 token, trying without token", "error", err)
		// Fallback: try without token (IMDSv1) for backward compatibility
		if impl.tryMetadataWithoutToken(client, detected) {
			return
		}
		detected <- bean.Unknown
		return
	}

	// Try with IMDSv2 token
	if impl.tryMetadataWithToken(client, token, detected) {
		return
	}

	detected <- bean.Unknown
}

// getIMDSv2Token gets the IMDSv2 token for metadata service access
func (impl *IdentifyAmazon) getIMDSv2Token(client *http.Client) (string, error) {
	req, err := http.NewRequest("PUT", bean.TokenForAmazonMetadataServerV2, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-aws-ec2-metadata-token-ttl-seconds", "21600")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", err
	}

	token, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(token), nil
}

// tryMetadataWithToken attempts to identify AWS using IMDSv2 with token
func (impl *IdentifyAmazon) tryMetadataWithToken(client *http.Client, token string, detected chan<- string) bool {
	req, err := http.NewRequest("GET", bean.AmazonMetadataServer, nil)
	if err != nil {
		impl.Logger.Debugw("Error creating metadata request", "error", err)
		return false
	}
	req.Header.Set("X-aws-ec2-metadata-token", token)

	return impl.processMetadataResponse(client, req, detected)
}

// tryMetadataWithoutToken attempts to identify AWS using IMDSv1 (fallback)
func (impl *IdentifyAmazon) tryMetadataWithoutToken(client *http.Client, detected chan<- string) bool {
	req, err := http.NewRequest("GET", bean.AmazonMetadataServer, nil)
	if err != nil {
		impl.Logger.Debugw("Error creating metadata request", "error", err)
		return false
	}

	return impl.processMetadataResponse(client, req, detected)
}

// processMetadataResponse processes the metadata service response and determines if it's AWS
func (impl *IdentifyAmazon) processMetadataResponse(client *http.Client, req *http.Request, detected chan<- string) bool {
	resp, err := client.Do(req)
	if err != nil {
		impl.Logger.Debugw("Error requesting metadata", "error", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		impl.Logger.Debugw("Metadata service returned non-200 status", "status", resp.StatusCode)
		return false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		impl.Logger.Debugw("Error reading metadata response", "error", err)
		return false
	}

	var r instanceIdentityResponse
	err = json.Unmarshal(body, &r)
	if err != nil {
		impl.Logger.Debugw("Error unmarshaling metadata response", "error", err, "body", string(body))
		return false
	}

	// Enhanced AWS detection logic for EKS compatibility
	isAWS := false

	// Traditional EC2 detection (backward compatibility)
	if strings.HasPrefix(r.ImageID, "ami-") && strings.HasPrefix(r.InstanceID, "i-") {
		impl.Logger.Debugw("AWS detected via traditional EC2 metadata", "imageId", r.ImageID, "instanceId", r.InstanceID)
		isAWS = true
	}

	// EKS/Fargate detection - check for AWS region format
	if r.Region != "" && impl.isValidAWSRegion(r.Region) {
		impl.Logger.Debugw("AWS detected via region metadata", "region", r.Region)
		isAWS = true
	}

	// Check availability zone format (AWS-specific)
	if r.AvailabilityZone != "" && impl.isValidAWSAvailabilityZone(r.AvailabilityZone) {
		impl.Logger.Debugw("AWS detected via availability zone", "az", r.AvailabilityZone)
		isAWS = true
	}

	if isAWS {
		detected <- bean.Amazon
		return true
	}

	return false
}

// isValidAWSRegion checks if the region follows AWS region naming convention
func (impl *IdentifyAmazon) isValidAWSRegion(region string) bool {
	// AWS regions follow pattern: us-east-1, eu-west-1, ap-southeast-2, etc.
	parts := strings.Split(region, "-")
	return len(parts) >= 3 && len(parts[len(parts)-1]) == 1
}

// isValidAWSAvailabilityZone checks if the AZ follows AWS AZ naming convention
func (impl *IdentifyAmazon) isValidAWSAvailabilityZone(az string) bool {
	// AWS AZs follow pattern: us-east-1a, eu-west-1b, etc.
	if len(az) < 4 {
		return false
	}
	// Should end with a single letter
	lastChar := az[len(az)-1]
	return lastChar >= 'a' && lastChar <= 'z'
}
