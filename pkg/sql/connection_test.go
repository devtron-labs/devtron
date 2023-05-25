//go:build unit
// +build unit

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

package sql

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

/*
func TestConnection(t *testing.T) {
	var n int
	_, err := dbConnection.QueryOne(pg.Scan(&n), "SELECT 1")
	assert.NoError(t, err, "error in db connection")
	assert.Equal(t, 1, n, "unexpected result from db")
}
func TestCloseConnection(t *testing.T) {
	err := closeConnection()
	assert.NoError(t, err, "error in closing connection")
}
*/

func TestObfuscation(t *testing.T) {
	cfg, _ := GetConfig()
	cfg1 := obfuscateSecretTags(cfg).(*Config)
	t.Log("log data", "cfg", cfg)
	t.Log("log data", "cfg1", cfg1)
	fmt.Printf("config found: %v\n", obfuscateSecretTags(cfg1))
	assert.Equal(t, cfg.Addr, cfg1.Addr)
	assert.Equal(t, cfg.User, cfg1.User)
	assert.Equal(t, cfg.ApplicationName, cfg1.ApplicationName)
	assert.Equal(t, cfg.Port, cfg1.Port)
	assert.NotEqual(t, cfg.Password, cfg1.Password)
}
