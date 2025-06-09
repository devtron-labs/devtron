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
	"encoding/json"
	"io"
)

// CompressJSON marshals and compresses JSON data
func CompressJSON(data interface{}) ([]byte, error) {
	// Marshal to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Compress using gzip
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)

	_, err = gzipWriter.Write(jsonData)
	if err != nil {
		return nil, err
	}

	err = gzipWriter.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// DecompressJSON decompresses and unmarshals JSON data
func DecompressJSON(compressedData []byte, target interface{}) error {
	// Decompress
	reader, err := gzip.NewReader(bytes.NewReader(compressedData))
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

// CompressString compresses string data
func CompressString(data string) ([]byte, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)

	_, err := gzipWriter.Write([]byte(data))
	if err != nil {
		return nil, err
	}

	err = gzipWriter.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// DecompressString decompresses string data
func DecompressString(compressedData []byte) (string, error) {
	reader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return "", err
	}
	defer reader.Close()

	decompressedData, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(decompressedData), nil
}

// Utility functions for backward compatibility

// CompressWorkflowRequest compresses WorkflowRequest to bytes
func CompressWorkflowRequest(workflowRequest interface{}) ([]byte, error) {
	return CompressJSON(workflowRequest)
}

// DecompressWorkflowRequest decompresses bytes to WorkflowRequest
func DecompressWorkflowRequest(compressedData []byte, target interface{}) error {
	return DecompressJSON(compressedData, target)
}

// Example usage:
/*
// Compress before saving
compressedData, err := CompressWorkflowRequest(workflowRequest)
if err != nil {
    return err
}
snapshot.WorkflowRequestJson = compressedData

// Decompress when reading
var workflowRequest types.WorkflowRequest
err = DecompressWorkflowRequest(snapshot.WorkflowRequestJson, &workflowRequest)
if err != nil {
    return err
}
*/
