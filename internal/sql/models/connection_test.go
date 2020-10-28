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

// +build unit

package models

import (
	"testing"

	// "devtron.io/front/internal/sql/models"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
)

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
