package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DbConfig struct {
	// The database names that are reserved and cannot be created.
	ReserveDbNames []string `mapstructure:"reserve_db_names" json:"reserve_db_names"`
	// The pg instance host.
	Host string `mapstructure:"host" json:"host"`
	// The pg instance port.
	Port int `mapstructure:"port" json:"port"`
	// The pg instance super user.
	User string `mapstructure:"user" json:"user"`
	// The default database use by super user
	DbName string `mapstructure:"db_name" json:"db_name"`
	// The password of the super user.
	Password_ string `mapstructure:"password" json:"password"`
	// The file save the password
	PasswordFile string `mapstructure:"password_file" json:"password_file"`
	// The max connections to the database.
	MaxConns int32 `mapstructure:"max_conns" json:"max_conns"`
	// The path of the database migrations.
	MigrationsPath string `mapstructure:"migrations_path" json:"migrations_path"`
}

func (c *DbConfig) Url() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", c.User, url.QueryEscape(c.Password()), c.Host, c.Port, c.DbName)
}

func (c *DbConfig) Password() string {
	if c.Password_ == "" {
		if c.PasswordFile != "" {
			password, err := os.ReadFile(c.PasswordFile)
			if err != nil {
				log.Fatal("Failed to read the password file, error: ", err)
			}
			c.Password_ = strings.TrimSpace(string(password))
		}
	}
	return c.Password_
}

func (c *DbConfig) NewPoolConfig() *pgxpool.Config {
	const defaultMinConns = int32(0)
	const defaultMaxConnLifetime = time.Hour
	const defaultMaxConnIdleTime = time.Minute * 30
	const defaultHealthCheckPeriod = time.Minute
	const defaultConnectTimeout = time.Second * 5

	dbConfig, err := pgxpool.ParseConfig(c.Url())
	if err != nil {
		detail := string(err.Error())
		detail = strings.ReplaceAll(detail, fmt.Sprintf(":%s@", url.QueryEscape(c.Password())), ":******@")
		log.Fatal("Failed to create a config, error: ", detail)
	}

	dbConfig.MaxConns = c.MaxConns
	dbConfig.MinConns = defaultMinConns
	dbConfig.MaxConnLifetime = defaultMaxConnLifetime
	dbConfig.MaxConnIdleTime = defaultMaxConnIdleTime
	dbConfig.HealthCheckPeriod = defaultHealthCheckPeriod
	dbConfig.ConnConfig.ConnectTimeout = defaultConnectTimeout

	return dbConfig
}
