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

package util

import "testing"

func TestAutoscale(t *testing.T) {
	type args struct {
		dat map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//got, err := Autoscale(tt.args.dat)
			//if (err != nil) != tt.wantErr {
			//	t.Errorf("Autoscale() error = %v, wantErr %v", err, tt.wantErr)
			//	return
			//}
			//if got != tt.want {
			//	t.Errorf("Autoscale() got = %v, want %v", got, tt.want)
			//}
		})
	}
}
