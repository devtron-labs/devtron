/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package authentication

type Config struct {
	AuthEnabled bool
}

func GetConfig() *Config {
	cfg := &Config{AuthEnabled: true}
	return cfg
}
