/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package helper

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"io"
)

// CompressWorkflowRequest compresses WorkflowRequest to bytes
func CompressWorkflowRequest(workflowRequest interface{}) (string, error) {
	// Marshal to JSON
	jsonData, err := json.Marshal(workflowRequest)
	if err != nil {
		return "", err
	}

	// Compress using gzip
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)

	_, err = gzipWriter.Write(jsonData)
	if err != nil {
		return "", err
	}

	err = gzipWriter.Close()
	if err != nil {
		return "", err
	}

	// Encode compressed binary data to Base64 to avoid UTF-8 encoding issues
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// DecompressWorkflowRequest decompresses bytes to WorkflowRequest
// This function handles both Base64 encoded and legacy raw binary data for backward compatibility
func DecompressWorkflowRequest(compressedData string, target interface{}) error {
	// Try Base64 decoding first (new format)
	decodedData, err := base64.StdEncoding.DecodeString(compressedData)
	if err != nil {
		return err
	}

	// Use decoded data for decompression
	reader, err := gzip.NewReader(bytes.NewReader(decodedData))
	if err != nil {
		return err
	}
	defer reader.Close()

	// Read decompressed data
	decompressedData, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	// Unmarshal JSON
	return json.Unmarshal(decompressedData, target)
}
