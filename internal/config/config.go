package config

import (
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

const (
	prefix = ""
)

type Config struct {
	Logger           Logger        `envconfig:"LOGGER"`
	APIServer        Server        `envconfig:"API_SERVER"`
	TechServer       Server        `envconfig:"TECH_SERVER"`
	Postgres         Postgres      `envconfig:"POSTGRES"`
	Mailer           Mailer        `envconfig:"MAILER"`
	MigrationsDir    string        `default:"./migrations"     envconfig:"MIGRATIONS_DIR"`
	FrontendURL      string        `envconfig:"FRONTEND_URL"   required:"true"`
	SecretKey        string        `envconfig:"SECRET_KEY"     required:"true"`
	JWTSecretKey     string        `envconfig:"JWT_SECRET_KEY" required:"true"`
	AccessTokenTTL   time.Duration `default:"3h"               envconfig:"ACCESS_TOKEN_TTL"`
	RefreshTokenTTL  time.Duration `default:"168h"             envconfig:"REFRESH_TOKEN_TTL"`
	ResetPasswordTTL time.Duration `default:"8h"               envconfig:"RESET_PASSWORD_TTL"`

	AdminEmail       string `envconfig:"ADMIN_EMAIL"`
	AdminTmpPassword string `envconfig:"ADMIN_TMP_PASSWORD"`

	// SSO Configuration
	// Keycloak KeycloakConfig `envconfig:"KEYCLOAK"`
	SAML SAMLConfig `envconfig:"SAML"`
}

type Logger struct {
	Lvl string `default:"info" envconfig:"LEVEL"`
}

func (l *Logger) Level() slog.Level {
	switch l.Lvl {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		panic("invalid logger level " + l.Lvl)
	}
}

type Server struct {
	Addr         string        `envconfig:"ADDR" required:"true"`
	ReadTimeout  time.Duration `default:"15s"    envconfig:"READ_TIMEOUT"`
	WriteTimeout time.Duration `default:"30s"    envconfig:"WRITE_TIMEOUT"`
	IdleTimeout  time.Duration `default:"60s"    envconfig:"IDLE_TIMEOUT"`
}

// SAMLConfig holds SAML configuration.
type SAMLConfig struct {
	Enabled          bool              `default:"false" envconfig:"ENABLED"`
	CreateCerts      bool              `default:"false" envconfig:"CREATE_CERTS"`
	EntityID         string            `default:""      envconfig:"ENTITY_ID"`
	CertificatePath  string            `default:""      envconfig:"CERTIFICATE_PATH"`
	PrivateKeyPath   string            `default:""      envconfig:"PRIVATE_KEY_PATH"`
	IDPMetadataURL   string            `default:""      envconfig:"IDP_METADATA_URL"`
	SSOURL           string            `default:""      envconfig:"SSO_URL"` // Optional: override SSO URL from metadata
	AttributeMapping map[string]string `default:""      envconfig:"ATTRIBUTE_MAPPING"`
	SkipTLSVerify    bool              `default:"false" envconfig:"SKIP_TLS_VERIFY"`
}

type Mailer struct {
	Addr          string `envconfig:"ADDR"     required:"true"`
	User          string `envconfig:"USER"     required:"true"`
	Password      string `envconfig:"PASSWORD" required:"true"`
	From          string `envconfig:"FROM"     required:"true"`
	AllowInsecure bool   `default:"false"      envconfig:"ALLOW_INSECURE"`
	CertFile      string `default:""           envconfig:"CERT_FILE"`
	KeyFile       string `default:""           envconfig:"KEY_FILE"`
	UseTLS        bool   `default:"false"      envconfig:"USE_TLS"`
}

type Postgres struct {
	User            string        `envconfig:"USER"     required:"true"`
	Password        string        `envconfig:"PASSWORD" required:"true"`
	Host            string        `envconfig:"HOST"     required:"true"`
	Port            string        `default:"5432"       envconfig:"PORT"`
	Database        string        `envconfig:"DATABASE" required:"true"`
	MaxIdleConnTime time.Duration `default:"5m"         envconfig:"MAX_IDLE_CONN_TIME"`
	MaxConns        int           `default:"20"         envconfig:"MAX_CONNS"`
	ConnMaxLifetime time.Duration `default:"10m"        envconfig:"CONN_MAX_LIFETIME"`
}

func (db *Postgres) ConnString() string {
	var user *url.Userinfo

	if db.User != "" {
		var pass string

		if db.Password != "" {
			pass = db.Password
		}

		user = url.UserPassword(db.User, pass)
	}

	params := url.Values{}
	params.Set("sslmode", "disable")
	params.Set("connect_timeout", "10")

	uri := url.URL{
		Scheme:   "postgres",
		User:     user,
		Host:     net.JoinHostPort(db.Host, db.Port),
		Path:     db.Database,
		RawQuery: params.Encode(),
	}

	return uri.String()
}

func (db *Postgres) ConnStringWithPoolSize() string {
	connString := db.ConnString()

	return connString + fmt.Sprintf("&pool_max_conns=%d", db.MaxConns)
}

func New(filePath string) (*Config, error) {
	cfg := &Config{}

	if filePath != "" {
		if err := godotenv.Load(filePath); err != nil {
			return nil, fmt.Errorf("error loading env file: %w", err)
		}
	}

	if err := envconfig.Process(prefix, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func MustNew(filePath string) *Config {
	cfg, err := New(filePath)
	if err != nil {
		panic(err)
	}

	return cfg
}
