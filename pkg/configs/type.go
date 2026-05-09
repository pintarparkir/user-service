package configs

// Config holds user-service configuration.
// All fields are loaded from environment variables with safe defaults.
type Config struct {
	AppName  string `env:"APP_NAME" envDefault:"user-service"`
	AppEnv   string `env:"APP_ENV" envDefault:"local"`
	AppPort  string `env:"APP_PORT" envDefault:"8080"`  // REST port (Cloud Run injects PORT; default 8080)
	GrpcPort string `env:"GRPC_PORT" envDefault:"9094"` // gRPC port for s2s callers (notification, reservation)

	// RS256 public key PEM from super-app. If empty, JWT signature check is skipped (dev only).
	SuperAppJWTPubKey string `env:"SUPER_APP_JWT_PUBLIC_KEY_PEM" envDefault:""`

	DbHost     string `env:"DB_HOST" envDefault:"localhost"`
	DbPort     string `env:"DB_PORT" envDefault:"5432"`
	DbUsername string `env:"DB_USERNAME" envDefault:"postgres"`
	DbPassword string `env:"DB_PASSWORD" envDefault:"postgres"`
	DbName     string `env:"DB_NAME" envDefault:"user_service"`
	DbMaxOpen  int    `env:"DB_MAX_OPEN" envDefault:"25"`
	DbMaxIdle  int    `env:"DB_MAX_IDLE" envDefault:"10"`

	RedisHost      string `env:"REDIS_HOST" envDefault:"localhost"`
	RedisPort      string `env:"REDIS_PORT" envDefault:"6379"`
	RedisPassword  string `env:"REDIS_PASSWORD" envDefault:""`
	RedisDB        int    `env:"REDIS_DB" envDefault:"0"`
	RedisAppConfig string `env:"REDIS_APP_CONFIG" envDefault:"user-service"`

	OTLPEndpoint string `env:"OTLP_ENDPOINT" envDefault:"localhost:4317"`

	// pgcrypto symmetric key for PII columns (phone_e164_enc, email_enc).
	// Source from Cloud Secret Manager in production — never commit a real key.
	PgCryptoKey string `env:"PG_CRYPTO_KEY" envDefault:"local-dev-pgcrypto-key-change-me"`

	// SMS sender ID shown on recipient handset. Stub client logs if empty.
	SmsSenderID string `env:"SMS_SENDER_ID" envDefault:"ParkirPintar"`
}

// ConfigLoader controls the source of config (env file path, etc.).
type ConfigLoader struct {
	Env     string // local, staging, production
	EnvFile string // optional path to .env
}
