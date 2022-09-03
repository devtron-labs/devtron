package bean

type Config struct {
	IgnoreAuthCheck bool `env:"IGNORE_AUTOCOMPLETE_AUTH_CHECK" envDefault:"false"`
}
