package agent

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type DbConfig struct {
	// The database names that are reserved and cannot be created.
	ReserveNames []string `default:"postgres,template0,template1"`

	// The pg instance host.
	HostTemplate string `default:"127.0.0.1" env:"PG_HELPER_DB_HOST_TEMPLATE"`
	// The pg instance port.
	Port int `default:"5432"`
	// The pg instance super user.
	User string `default:"postgres"`
	// The default database use by super user
	Name string `default:"postgres"`
	// The password of the super user.
	Password_ string `env:"PG_HELPER_DB_PASSWORD" name:"password"`
	// The file save the password
	PasswordFile string `env:"PG_HELPER_DB_PASSWORD_FILE"`
	// The max connections to the database.
	MaxConns int32 `default:"4"`

	// The path of the database backups.
	BackupRootPath string `default:"/var/lib/pg-helper/backups"`
	// The majar version of the database that pg-helper work with.
	CurrentVersion int32 `env:"PG_MAJOR"`
}

func (c *DbConfig) Host(pg_version int32) string {
	if strings.Contains(c.HostTemplate, "%") {
		return fmt.Sprintf(c.HostTemplate, pg_version)
	}
	return c.HostTemplate
}

func (c *DbConfig) Url(dbName string, pg_version int32) string {
	if dbName == "" {
		dbName = c.Name
	}
	if pg_version == 0 {
		pg_version = c.CurrentVersion
	}
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable", c.User, url.QueryEscape(c.Password()), c.Host(pg_version), c.Port, dbName)
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

	dbConfig, err := pgxpool.ParseConfig(c.Url("", 0))
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
	return fmt.Sprintf("%s/pg-%d/%s", c.BackupRootPath, c.CurrentVersion, dbName)
}

// The backup file is relative to the BackupRootPath
func (c *DbConfig) NewBackupFile(dbName string) string {
	return fmt.Sprintf("pg-%d/%s/%s.sql", c.CurrentVersion, dbName, time.Now().Format("2006-01-02_15:04:05"))
}

func (c *DbConfig) extractBackupPath(backupPath string) (dbName string, pgVersion int32, err error) {
	dbName = ""
	pgVersion = 0

	cleanPath := filepath.Clean(backupPath)
	if !strings.HasPrefix(cleanPath, "pg-") {
		err = fmt.Errorf("illegal backup path")
		return
	}

	dbName = strings.Split(cleanPath, "/")[1]
	v, err := strconv.Atoi(strings.Split(cleanPath, "/")[0][3:])
	if err != nil {
		err = fmt.Errorf("illegal backup path, can not parse pg version")
	}
	pgVersion = int32(v)
	return
}

func (c *DbConfig) ValidateBackupPath(backupPath string, dbName string) (pgVersionInPath int32, err error) {
	dbNameInPath, pgVersionInPath, err := c.extractBackupPath(backupPath)
	if err != nil {
		return
	}

	if dbNameInPath != dbName {
		err = fmt.Errorf("illegal backup path, db name not match")
		return
	}

	if pgVersionInPath > c.CurrentVersion {
		err = fmt.Errorf("illegal backup path, can not restore from a higher version")
		return
	}

	return
}

func (c *DbConfig) IsReservedName(name string) bool {
	for _, n := range c.ReserveNames {
		if n == name {
			return true
		}
	}
	return false
}