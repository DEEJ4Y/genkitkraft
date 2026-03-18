# Configuration Management

Keep configuration centralized and loaded from environment variables:

```go
// internal/config/config.go

type Config struct {
    Env        string
    ReadDB     DBConfig
    WriteDB    DBConfig
    GRPC       GRPCConfig
    Cache      CacheConfig
    Blob       BlobConfig
    Auth       AuthConfig
}

type DBConfig struct {
    Host     string
    Port     int
    Name     string
    User     string
    Password string
}

func Load() Config {
    return Config{
        Env: os.Getenv("APP_ENV"),
        ReadDB: DBConfig{
            Host:     os.Getenv("READ_DB_HOST"),
            Port:     getEnvInt("READ_DB_PORT", 3306),
            Name:     os.Getenv("READ_DB_NAME"),
            User:     os.Getenv("READ_DB_USER"),
            Password: os.Getenv("READ_DB_PASSWORD"),
        },
        // ...
    }
}

func getEnvInt(key string, fallback int) int {
    val := os.Getenv(key)
    if val == "" {
        return fallback
    }
    n, err := strconv.Atoi(val)
    if err != nil {
        return fallback
    }
    return n
}
```

## Rules

- Configuration structs live in `internal/config/`.
- All configuration is read from environment variables — no config files.
- Configuration is loaded once at startup and injected through the composition root.
- Adapters receive only the configuration they need — not the entire config.
