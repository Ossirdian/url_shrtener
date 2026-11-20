package config

type LoggerConfig interface {
	Level() string
	AsJSON() bool
}

type HTTPConfig interface {
	Addr() string
}

type AuthConfig interface {
	Username() string
	Password() string
}

type PostgresConfig interface {
	DSN() string
	MigrationsDir() string
	AutoMigrate() bool
}

type Config interface {
	Logger() LoggerConfig
	HTTP() HTTPConfig
	Auth() AuthConfig
	Postgres() PostgresConfig
}
