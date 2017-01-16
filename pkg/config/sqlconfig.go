package config

import ()

// SnmpDeviceCfg contains all snmp related device definitions
type SnmpDeviceCfg struct {
	ID string `xorm:"'id' unique"`
	//snmp connection config
	Host    string `xorm:"host"`
	Port    int    `xorm:"port"`
	Retries int    `xorm:"retries"`
	Timeout int    `xorm:"timeout"`
	Repeat  int    `xorm:"repeat"`
	Active  bool   `xorm:"'active' default 1"`
	//snmp auth  config
	SnmpVersion string `xorm:"snmpversion"`
	Community   string `xorm:"community"`
	V3SecLevel  string `xorm:"v3seclevel"`
	V3AuthUser  string `xorm:"v3authuser"`
	V3AuthPass  string `xorm:"v3authpass"`
	V3AuthProt  string `xorm:"v3authprot"`
	V3PrivPass  string `xorm:"v3privpass"`
	V3PrivProt  string `xorm:"v3privprot"`
	//snmp workarround for some devices
	DisableBulk bool `xorm:"'disablebulk' default 0"`
	//snmp runtime config
	Freq          int `xorm:"'freq' default 60"`
	UpdateFltFreq int `xorm:"'update_flt_freq' default 60"`

	OutDB    string `xorm:"outdb"`
	LogLevel string `xorm:"loglevel"`
	LogFile  string `xorm:"logfile"`

	SnmpDebug bool `xorm:"snmpdebug"`
	//influx tags
	DeviceTagName  string   `xorm:"devicetagname"`
	DeviceTagValue string   `xorm:"devicetagvalue"`
	ExtraTags      []string `xorm:"extra-tags"`
	Description    string   `xorm:"description"`

	//Filters for measurements
	MeasurementGroups []string `xorm:"-"`
	MeasFilters       []string `xorm:"-"`
}

// InfluxCfg is the main configuration for any InfluxDB TSDB
type InfluxCfg struct {
	ID          string `xorm:"'id' unique"`
	Host        string `xorm:"host"`
	Port        int    `xorm:"port"`
	DB          string `xorm:"db"`
	User        string `xorm:"user"`
	Password    string `xorm:"password"`
	Retention   string `xorm:"retention"`
	Timeout     int    `xorm:"'timeout' default 30"`
	UserAgent   string `xorm:"useragent"`
	Description string `xorm:"description"`
}

//MeasFilterCfg the filter configuration
type MeasFilterCfg struct {
	ID               string `xorm:"'id' unique"`
	IDMeasurementCfg string `xorm:"id_measurement_cfg"`
	FType            string `xorm:"filter_type"`       //file/OIDCondition/CustomFilter
	FileName         string `xorm:"file_name"`         //only valid if  type=file
	CustomID         string `xorm:"'customid' unique"` //only valid if type = CustomFilter
	EnableAlias      bool   `xorm:"enable_alias"`      //only valid if file/CustomFilter
	OIDCond          string `xorm:"cond_oid"`
	CondType         string `xorm:"cond_type"`
	CondValue        string `xorm:"cond_value"`
	Description      string `xorm:"description"`
}

//MeasurementFieldCfg the metrics contained on each measurement (to initialize on the fieldMetric array)
type MeasurementFieldCfg struct {
	IDMeasurementCfg string `xorm:"id_measurement_cfg"`
	IDMetricCfg      string `xorm:"id_metric_cfg"`
	Report           bool   `xorm:"'report' default 1"`
}

// CustomFilterCfg table with user custom choosed indexes
type CustomFilterCfg struct {
	customid string `xorm:"customid"`
	tagid    string `xorm:"tagid"`
	alias    string `xorm:"alias"`
}

// OidConditionCfg condition config for filters and metrics
type OidConditionCfg struct {
	ID          string `xorm:"'id' unique"`
	OIDCond     string `xorm:"cond_oid"`
	CondType    string `xorm:"cond_type"`
	CondValue   string `xorm:"cond_value"`
	Description string `xorm:"description"`
}

//SnmpDevFilters filters to use with indexed measurement
type SnmpDevFilters struct {
	IDSnmpDev string `xorm:"id_snmpdev"`
	IDFilter  string `xorm:"id_filter"`
}

//MGroupsCfg measurement groups to asign to devices
type MGroupsCfg struct {
	ID           string   `xorm:"'id' unique"`
	Measurements []string `xorm:"-"`
	Description  string   `xorm:"description"`
}

//MGroupsMeasurements measurements contained on each Measurement Group
type MGroupsMeasurements struct {
	IDMGroupCfg      string `xorm:"id_mgroup_cfg"`
	IDMeasurementCfg string `xorm:"id_measurement_cfg"`
}

// SnmpDevMGroups Mgroups defined on each SnmpDevice
type SnmpDevMGroups struct {
	IDSnmpDev   string `xorm:"id_snmpdev"`
	IDMGroupCfg string `xorm:"id_mgroup_cfg"`
}

// SQLConfig read from DB
type SQLConfig struct {
	Metrics      map[string]*SnmpMetricCfg
	Measurements map[string]*MeasurementCfg
	MFilters     map[string]*MeasFilterCfg
	GetGroups    map[string]*MGroupsCfg
	SnmpDevice   map[string]*SnmpDeviceCfg
	Influxdb     map[string]*InfluxCfg
}

/*
initMetricsCfg this function does 2 things
1.- Initialice id from key of maps for all SnmpMetricCfg and InfluxMeasurementCfg objects
2.- Initialice references between InfluxMeasurementCfg and SnmpMetricGfg objects
*/
// InitMetricsCfg xx
func InitMetricsCfg(cfg *SQLConfig) error {
	//TODO:
	// - check duplicates OID's => warning messages
	//Initialize references to SnmpMetricGfg into InfluxMeasurementCfg
	log.Debug("--------------------Initializing Config metrics-------------------")
	log.Debug("Initializing SNMPMetricconfig...")
	for mKey, mVal := range cfg.Metrics {
		err := mVal.Init()
		if err != nil {
			log.Warnln("Error in Metric config:", err)
			//if some error int the format the metric is deleted from the config
			delete(cfg.Metrics, mKey)
		}
	}
	log.Debug("Initializing MEASSUREMENTSconfig...")
	for mKey, mVal := range cfg.Measurements {
		err := mVal.Init(&cfg.Metrics)
		if err != nil {
			log.Warnln("Error in Measurement config:", err)
			//if some error int the format the metric is deleted from the config
			delete(cfg.Metrics, mKey)
		}

		log.Debugf("FIELDMETRICS: %+v", mVal.FieldMetric)
	}
	log.Debug("-----------------------END Config metrics----------------------")
	return nil
}

//var DBConfig SQLConfig
