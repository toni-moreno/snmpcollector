package config

import (
	"github.com/go-xorm/xorm"
)

// GeneralConfig has miscellaneous configuration options
type GeneralConfig struct {
	InstanceID string `toml:"instanceID"`
	LogDir     string `toml:"logdir"`
	HomeDir    string `toml:"homedir"`
	DataDir    string `toml:"datadir"`
	LogLevel   string `toml:"loglevel"`
}

//DatabaseCfg de configuration for the database
type DatabaseCfg struct {
	numChanges int64  `toml:"-"`
	Type       string `toml:"type"`
	Host       string `toml:"host"`
	Name       string `toml:"name"`
	User       string `toml:"user"`
	Password   string `toml:"password"`
	SQLLogFile string `toml:"sqllogfile"`
	Debug      string `toml:"debug"`
	x          *xorm.Engine
}

//SelfMonConfig configuration for self monitoring
type SelfMonConfig struct {
	Enabled           bool     `toml:"enabled"`
	Freq              int      `toml:"freq"`
	Prefix            string   `toml:"prefix"`
	InheritDeviceTags bool     `toml:"inheritdevicetags"`
	ExtraTags         []string `toml:"extra-tags"`
}

//HTTPConfig has webserver config options
type HTTPConfig struct {
	Port          int    `toml:"port"`
	AdminUser     string `toml:"adminuser"`
	AdminPassword string `toml:"adminpassword"`
	CookieID      string `toml:"cookieid"`
}

//Config Main Configuration struct
type Config struct {
	General  GeneralConfig
	Database DatabaseCfg
	Selfmon  SelfMonConfig
	HTTP     HTTPConfig
}

//var MainConfig Config
