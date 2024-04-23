package server

type DbConfig struct {
	MandatoryPgVersions []int32 `help:"The PostgreSQL versions that must running"`
}
