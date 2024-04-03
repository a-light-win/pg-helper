package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
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

	// The path of the database backups.
	BackupRootPath string `mapstructure:"backup_root_path" json:"backup_root_path"`
	// The majar version of the database that pg-helper work with.
	CurrentDbVersion int `mapstructure:"current_db_version" json:"current_db_version"`
}

func (c *DbConfig) Url(dbName string) string {
	if dbName == "" {
		dbName = c.DbName
	}
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable", c.User, url.QueryEscape(c.Password()), c.Host, c.Port, dbName)
}

func (c *DbConfig) Password() string {
	if c.Password_ == "" {
		if c.PasswordFile != "" {
			password, err := os.ReadFile(c.PasswordFile)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to read the password file")
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

	dbConfig, err := pgxpool.ParseConfig(c.Url(""))
	if err != nil {
		detail := string(err.Error())
		detail = strings.ReplaceAll(detail, fmt.Sprintf(":%s@", url.QueryEscape(c.Password())), ":******@")
		log.Fatal().Str("Error", detail).Msg("Failed to create a pool config")
	}

	dbConfig.MaxConns = c.MaxConns
	dbConfig.MinConns = defaultMinConns
	dbConfig.MaxConnLifetime = defaultMaxConnLifetime
	dbConfig.MaxConnIdleTime = defaultMaxConnIdleTime
	dbConfig.HealthCheckPeriod = defaultHealthCheckPeriod
	dbConfig.ConnConfig.ConnectTimeout = defaultConnectTimeout

	return dbConfig
}

func (c *DbConfig) BackupDbDir(dbName string) string {
	return fmt.Sprintf("%s/pg-%d/%s", c.BackupRootPath, c.CurrentDbVersion, dbName)
}

// The backup file is relative to the BackupRootPath
func (c *DbConfig) NewBackupFile(dbName string) string {
	return fmt.Sprintf("pg-%d/%s/%s.sql", c.CurrentDbVersion, dbName, time.Now().Format("2006-01-02_15:04:05"))
}
