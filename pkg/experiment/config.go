package experiment

type Config struct {
	Debug              bool
	ServerUrl          string
	FetchTimeoutMillis int
	RetryBackoff       *RetryBackoff
}

type RetryBackoff struct {
	FetchRetries               int
	FetchRetryBackoffMinMillis int
	FetchRetryBackoffMaxMillis int
	FetchRetryBackoffScalar    float32
	FetchRetryTimeoutMillis    int
}

var DefaultRetryBackoff = &RetryBackoff{
	FetchRetries:               8,
	FetchRetryBackoffMinMillis: 500,
	FetchRetryBackoffMaxMillis: 10000,
	FetchRetryBackoffScalar:    1.5,
	FetchRetryTimeoutMillis:    10000,
}

var DefaultConfig = &Config{
	Debug:              false,
	ServerUrl:          "https://api.lab.amplitude.com/",
	FetchTimeoutMillis: 10000,
	RetryBackoff:       DefaultRetryBackoff,
}

func fillConfigDefaults(c *Config) *Config {
	if c == nil {
		return DefaultConfig
	}
	if c.ServerUrl == "" {
		c.ServerUrl = DefaultConfig.ServerUrl
	}
	if c.FetchTimeoutMillis == 0 {
		c.FetchTimeoutMillis = DefaultConfig.FetchTimeoutMillis
	}
	if c.RetryBackoff == nil {
		c.RetryBackoff = DefaultConfig.RetryBackoff
	}
	return c
}
