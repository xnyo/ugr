package common

// Configuration holds the bot configuration
type Configuration struct {
	Token        string `env:"TOKEN"`
	Username     string `env:"USERNAME"`
	DbDriver     string `env:"DB_DRIVER" env-default:"sqlite3" env-description:"sqlite3 or mysql"`
	DbDSN        string `env:"DB_DSN" env-default:"ugr.db"`
	LogChannelID string `env:"LOG_CHANNEL_ID"`
	SentryDSN    string `env:"SENTRY_DSN"`
	Debug        bool   `env:"DEBUG" env-default:"false"`
	Timezone     string `env:"TIMEZONE" env-default:"UTC" env-description:"Timezone for input/output. UTC is used internally."`
}
