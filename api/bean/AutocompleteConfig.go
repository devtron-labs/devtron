/*
 * Copyright (c) 2024. Devtron Inc.
 */

package bean

type Config struct {
	IgnoreAuthCheck bool `env:"IGNORE_AUTOCOMPLETE_AUTH_CHECK" envDefault:"false"`
}
