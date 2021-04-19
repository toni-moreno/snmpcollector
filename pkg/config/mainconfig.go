package config

import (
	"github.com/go-xorm/xorm"
)

// GeneralConfig has miscellaneous configuration options
type GeneralConfig struct {
	InstanceID  string `mapstructure:"instanceID" envconfig:"SNMPCOL_GENERAL_INSTANCE_ID"`
	LogDir      string `mapstructure:"logdir" envconfig:"SNMPCOL_GENERAL_LOG_DIR"`
	HomeDir     string `mapstructure:"homedir" envconfig:"SNMPCOL_GENERAL_HOME_DIR"`
	DataDir     string `mapstructure:"datadir" envconfig:"SNMPCOL_GENERAL_DATA_DIR" `
	LogLevel    string `mapstructure:"loglevel" envconfig:"SNMPCOL_GENERAL_LOG_LEVEL"`
	LogMode     string `mapstructure:"log_mode" envconfig:"SNMPCOL_GENERAL_LOG_MODE"`
	Location    string `mapstructure:"location" envconfig:"SNMPCOL_GENERAL_LOCATION"`
	Description string `mapstructure:"description" envconfig:"SNMPCOL_GENERAL_DESCRIPTION"`
}

//DatabaseCfg de configuration for the database
type DatabaseCfg struct {
	numChanges     int64  `mapstructure:"-" `
	Type           string `mapstructure:"type" envconfig:"SNMPCOL_DATABASE_DRIVER_TYPE"`
	Host           string `mapstructure:"host" envconfig:"SNMPCOL_DATABASE_SERVER_HOST"`
	Name           string `mapstructure:"name" envconfig:"SNMPCOL_DATABASE_NAME"`
	User           string `mapstructure:"user" envconfig:"SNMPCOL_DATABASE_USERNAME"`
	Password       string `mapstructure:"password" envconfig:"SNMPCOL_DATABASE_PASSWORD"`
	SslMode        string `mapstructure:"ssl_mode" envconfig:"SNMPCOL_DATABASE_SSL_MODE"`
	ClientCertPath string `mapstructure:"client_cert_path" envconfig:"SNMPCOL_DATABASE_CLIENT_CERT_PATH"`
	ClientKeyPath  string `mapstructure:"client_key_path" envconfig:"SNMPCOL_DATABASE_CLIENT_KEY_PATH"`
	CaCertPath     string `mapstructure:"ca_cert_path" envconfig:"SNMPCOL_DATABASE_CA_CERT_PATH"`
	SQLLogFile     string `mapstructure:"sqllogfile" envconfig:"SNMPCOL_DATABASE_SQL_LOG_FILE"`
	Debug          string `mapstructure:"debug" envconfig:"SNMPCOL_DATABASE_SQL_DEBUG"`
	LogMode        string `mapstructure:"log_mode" envconfig:"SNMPCOL_DATABASE_LOG_MODE"`
	x              *xorm.Engine
}

//SelfMonConfig configuration for self monitoring
type SelfMonConfig struct {
	Enabled           bool     `mapstructure:"enabled" envconfig:"SNMPCOL_SELFMON_ENABLED"`
	Freq              int      `mapstructure:"freq" envconfig:"SNMPCOL_SELFMON_FREQ"`
	Prefix            string   `mapstructure:"prefix" envconfig:"SNMPCOL_SELFMON_PREFIX"`
	InheritDeviceTags bool     `mapstructure:"inheritdevicetags" envconfig:"SNMPCOL_SELFMON_INHERIT_DEVICE_TAGS"`
	ExtraTags         []string `mapstructure:"extratags" envconfig:"SNMPCOL_SELFMON_EXTRATAGS"`
}

//HTTPConfig has webserver config options
// Port should be deprecated from version >= 0.8.1
type HTTPConfig struct {
	Port          int    `mapstructure:"port"  envconfig:"SNMPCOL_HTTP_PORT"`
	Listen        string `mapstructure:"listen"  envconfig:"SNMPCOL_HTTP_LISTEN"`
	Protocol      string `mapstructure:"protocol" envconfig:"SNMPCOL_HTTP_PROTOCOL"`
	CertKey       string `mapstructure:"cert_key" envconfig:"SNMPCOL_HTTP_CERT_KEY"`
	CertFile      string `mapstructure:"cert_file" envconfig:"SNMPCOL_HTTP_CERT_FILE"`
	AdminUser     string `mapstructure:"adminuser" envconfig:"SNMPCOL_HTTP_ADMIN_USER"`
	AdminPassword string `mapstructure:"adminpassword" envconfig:"SNMPCOL_HTTP_ADMIN_PASSWORD"`
	CookieID      string `mapstructure:"cookieid" envconfig:"SNMPCOL_HTTP_COOKIE_ID"`
}

//Config Main Configuration struct
type Config struct {
	General  GeneralConfig `mapstructure:"general"`
	Database DatabaseCfg   `mapstructure:"database"`
	Selfmon  SelfMonConfig `mapstructure:"selfmon"`
	HTTP     HTTPConfig    `mapstructure:"http"`
}

//var MainConfig Config
