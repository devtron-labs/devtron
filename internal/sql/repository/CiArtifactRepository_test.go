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

package repository

import (
	"fmt"
	"testing"
)

func TestMaterialInfo_Parse(t *testing.T) {
	tests := []struct {
		name    string
		info    string
		wantErr bool
	}{
		{
			name:    "single",
			wantErr: false,
			info: `[
  {
    "material": {
      "git-configuration": {
        "shallow-clone": false,
        "branch": "master",
        "url": "https://github.com/gocd-demo/node-bulletin-board.git"
      },
      "type": "git"
    },
    "changed": false,
    "modifications": [
      {
        "revision": "992382abb91a664b751cd5d2a6eb154915fcd6aa",
        "modified-time": "Jan 16, 2019 3:22:47 PM",
        "data": {}
      }
    ]
  }
]`,
		},
		{
			name:    "multi",
			wantErr: false,
			info: `[
    {
      "material": {
        "git-configuration": {
          "shallow-clone": false,
          "branch": "master",
          "url": "https://github.com/gocd-demo/node-bulletin-board.git"
        },
        "type": "git"
      },
      "changed": false,
      "modifications": [
        {
          "revision": "992382abb91a664b751cd5d2a6eb154915fcd6aa",
          "modified-time": "Jan 16, 2019 3:22:47 PM",
          "data": {}
        }
      ]
    },
    {
      "material": {
        "plugin-id": "git.fb",
        "scm-configuration": {
          "url": "https://github.com/kumarnishant/dem-app.git",
          "defaultBranch": "master",
          "branchwhitelist": "dev*"
        },
        "type": "scm"
      },
      "changed": true,
      "modifications": [
        {
          "revision": "92235640cb6aad48164eeda37b108a8f45d095d7",
          "modified-time": "Apr 26, 2019 9:20:07 AM",
          "nrevision":"abc",
          "data": {
            "CURRENT_BRANCH": "master"
          }
        }
      ]
    }
  ]`,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mi := &CiArtifact{MaterialInfo: tt.info, DataSource: "GOCD"}
			got, err := mi.ParseMaterialInfo()
			if (err != nil) != tt.wantErr {
				t.Errorf("MaterialInfo.ParseGocdInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println(got)
		})
	}
}
