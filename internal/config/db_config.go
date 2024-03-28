package config

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
	Password string `mapstructure:"password" json:"password"`
	// The file save the password
	PasswordFile string `mapstructure:"password_file" json:"password_file"`
}
