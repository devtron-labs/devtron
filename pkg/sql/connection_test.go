/*
 * Copyright (c) 2024. Devtron Inc.
 */

//go:build unit
// +build unit

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
