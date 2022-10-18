package config

// SnmpDeviceCfg contains all snmp related device definitions
// swagger:model SnmpDeviceCfg
type SnmpDeviceCfg struct {
	ID string `xorm:"'id' unique" binding:"Required"`
	// snmp connection config
	Host       string   `xorm:"host" binding:"Required;AlphaDashDot"`
	Port       int      `xorm:"port" binding:"Required"`
	SystemOIDs []string `xorm:"systemoids"` // for non MIB-2 based devices
	Retries    int      `xorm:"retries"`
	Timeout    int      `xorm:"timeout"`
	Repeat     int      `xorm:"repeat"`
	Active     bool     `xorm:"'active' default TRUE"`
	// snmp auth  config
	SnmpVersion       string `xorm:"snmpversion" binding:"Required;In(1,2c,3)"`
	Community         string `xorm:"community"`
	V3SecLevel        string `xorm:"v3seclevel"`
	V3AuthUser        string `xorm:"v3authuser"`
	V3AuthPass        string `xorm:"v3authpass"`
	V3AuthProt        string `xorm:"v3authprot"`
	V3PrivPass        string `xorm:"v3privpass"`
	V3PrivProt        string `xorm:"v3privprot"`
	V3ContextEngineID string `xorm:"v3contextengineid"`
	V3ContextName     string `xorm:"v3contextname"`
	// snmp workarround for some devices
	DisableBulk    bool  `xorm:"'disablebulk' default 0"`
	MaxRepetitions uint8 `xorm:"'maxrepetitions' default 50" binding:"Default(50);IntegerNotZero"`
	MaxOids        int   `xorm:"'maxoids' default 60"`
	// snmp runtime config
	Freq             int  `xorm:"'freq' default 60" binding:"Default(60);IntegerNotZero"`
	UpdateFltFreq    int  `xorm:"'update_flt_freq' default 60" binding:"Default(60);UIntegerAndLessOne"`
	ConcurrentGather bool `xorm:"'concurrent_gather' default 1"`

	OutDB    string `xorm:"outdb"`
	LogLevel string `xorm:"loglevel" binding:"Default(info)"`
	LogFile  string `xorm:"logfile"`

	SnmpDebug bool `xorm:"'snmpdebug' default 0"`
	// influx tags
	DeviceTagName  string   `xorm:"devicetagname" binding:"Default(hostname)"`
	DeviceTagValue string   `xorm:"devicetagvalue" binding:"Default(id)"`
	ExtraTags      []string `xorm:"extra-tags"`
	DeviceVars     []string `xorm:"devicevars"`
	Description    string   `xorm:"description"`
	// Filters for measurements
	MeasurementGroups []string `xorm:"-"`
	MeasFilters       []string `xorm:"-"`
}

// OutputCfg is the main configuration for any supported backend type
// the relation will be 1:1 to the backend ID
// swagger:model OutputCfg

type OutputCfg struct {
	ID              string `xorm:"'id' unique" binding:"Required"`
	BackendType     string `xorm:"-" binding:"Required;In(kafka,influxdb)"`
	Active          bool   `xorm:"'active' default TRUE"`
	EnqueueOnError  bool   `xorm:"'enqueue_on_error' default TRUE"`
	BufferSize      int    `xorm:"'buffer_size' default 131070" binding:"Required;IntegerNotZero"`
	MetricBatchSize int    `xorm:"'metric_batch_size' default 13000" binding:"Required;IntegerNotZero"`
	FlushInterval   int    `xorm:"'flush_interval' default 60" binding:"Required;IntegerNotZero"`
	Backend         string `xorm:"-"`
	Description     string `xorm:"description"`
}

// KafkaCfg is the main configuration for any Kafka
// swagger:model KafkaCfg
type KafkaCfg struct {
	ID                  string   `xorm:"'id' unique" binding:"Required"`
	Brokers             []string `xorm:"brokers" binding:"Required"`
	Topic               string   `xorm:"topic"`
	Method              string   `xorm:"topic_suffix_method" binding:"Required;In(measurement,tags)"`
	Keys                []string `xorm:"topic_suffix_keys"`
	Separator           string   `xorm:"topic_suffix_separator"`
	ExcludeTopicTag     bool     `xorm:"exclude_topic_tag"`
	TopicTag            string   `xorm:"topic_tag"`
	RoutingTag          string   `xorm:"routing_tag"`
	RoutingKey          string   `xorm:"routing_key"`
	Description         string   `xorm:"description"`
	RequiredAcks        int      `xorm:"'required_acks' default -1"`
	MaxRetry            int      `xorm:"'max_retry' default 3"`
	MaxMessageBytes     int      `xorm:"max_message_bytes"`
	IdempotentWrites    bool     `xorm:"idempotent_writes"`
	TLSCA               string   `xorm:"tls_ca"`
	TLSCert             string   `xorm:"tls_cert"`
	TLSKey              string   `xorm:"tls_key"`
	TLSKeyPwd           string   `xorm:"tls_key_pwd"`
	TLSMinVersion       string   `xorm:"tls_min_version"`
	InsecureSkipVerify  bool     `xorm:"insecure_skip_verify"`
	ServerName          string   `xorm:"tls_server_name"`
	Version             string   `xorm:"version"`
	ClientID            string   `xorm:"client_id"`
	CompressionCodec    int      `xorm:"compression_codec"`
	EnableTLS           bool     `xorm:"enable_tls"`
	Socks5ProxyEnabled  bool     `xorm:"socks5_enabled"`
	Socks5ProxyAddress  string   `xorm:"socks5_address"`
	Socks5ProxyUsername string   `xorm:"socks5_username"`
	Socks5ProxyPassword string   `xorm:"socks5_password"`
	MetadataFull        bool     `xorm:"metadata_full"`
}

// InfluxCfg is the main configuration for any InfluxDB TSDB
// swagger:model InfluxCfg
type InfluxCfg struct {
	ID                 string `xorm:"'id' unique" binding:"Required"`
	Host               string `xorm:"host" binding:"Required"`
	Port               int    `xorm:"port" binding:"Required;IntegerNotZero"`
	DB                 string `xorm:"db" binding:"Required"`
	User               string `xorm:"user" binding:"Required"`
	Password           string `xorm:"password" binding:"Required"`
	Retention          string `xorm:"'retention' default 'autogen'" binding:"Required"`
	Precision          string `xorm:"'precision' default 's'" binding:"Default(s);OmitEmpty;In(h,m,s,ms,u,ns)"` // posible values [h,m,s,ms,u,ns] default seconds for the nature of data
	Timeout            int    `xorm:"'timeout' default 30" binding:"Default(30);IntegerNotZero"`
	UserAgent          string `xorm:"useragent" binding:"Default(snmpcollector)"`
	EnableSSL          bool   `xorm:"enable_ssl"`
	SSLCA              string `xorm:"ssl_ca"`
	SSLCert            string `xorm:"ssl_cert"`
	SSLKey             string `xorm:"ssl_key"`
	InsecureSkipVerify bool   `xorm:"insecure_skip_verify"`
	BufferSize         int    `xorm:"-" json:"-"`
	Description        string `xorm:"description"`
}

// MeasFilterCfg the filter configuration
// swagger:model MeasFilterCfg
type MeasFilterCfg struct {
	ID               string `xorm:"'id' unique" binding:"Required"`
	IDMeasurementCfg string `xorm:"id_measurement_cfg"`
	FType            string `xorm:"filter_type" binding:"Required"` // file/OIDCondition/CustomFilter
	FilterName       string `xorm:"filter_name" binding:"Required"` // valid identificator for the filter depending on the type
	EnableAlias      bool   `xorm:"enable_alias"`                   // only valid if file/Custom
	Description      string `xorm:"description"`
}

// MeasurementFieldCfg the metrics contained on each measurement (to initialize on the fieldMetric array)
type MeasurementFieldCfg struct {
	IDMeasurementCfg string `xorm:"id_measurement_cfg"`
	IDMetricCfg      string `xorm:"id_metric_cfg"`
	Report           int    `xorm:"'report' default 1"`
}

// CUSTOM FILTER TYPES

// CustomFilterItems  list of items on each custom filter
type CustomFilterItems struct {
	CustomID string `xorm:"customid"`
	TagID    string `xorm:"tagid"`
	Alias    string `xorm:"alias"`
}

// CustomFilterCfg table with user custom choosed indexes
// swagger:model CustomFilterCfg
type CustomFilterCfg struct {
	ID          string `xorm:"'id' unique" binding:"Required"`
	Description string `xorm:"description"`
	RelatedDev  string `xorm:"related_dev"`
	RelatedMeas string `xorm:"related_meas"`
	Items       []struct {
		TagID string
		Alias string
	} `xorm:"-"`
}

// SnmpDevFilters filters to use with indexed measurement
type SnmpDevFilters struct {
	IDSnmpDev string `xorm:"id_snmpdev"`
	IDFilter  string `xorm:"id_filter"`
}

// MGroupsCfg measurement groups to assign to devices
// swagger:model MGroupsCfg
type MGroupsCfg struct {
	ID           string   `xorm:"'id' unique" binding:"Required"`
	Measurements []string `xorm:"-"`
	Description  string   `xorm:"description"`
}

// MGroupsMeasurements measurements contained on each Measurement Group
type MGroupsMeasurements struct {
	IDMGroupCfg      string `xorm:"id_mgroup_cfg"`
	IDMeasurementCfg string `xorm:"id_measurement_cfg"`
}

// SnmpDevMGroups Mgroups defined on each SnmpDevice
type SnmpDevMGroups struct {
	IDSnmpDev   string `xorm:"id_snmpdev"`
	IDMGroupCfg string `xorm:"id_mgroup_cfg"`
}

// create unique entry based on output_id+backend_id,backend_type so we allow to define multiple
// backends with the same name and they differ on backend_type
// backend_type allow us to distinguish between different supported backends
// and allow them to be extended with custom ui
// even if we support multiple outputs, the relations are still ok
type OutputBackends struct {
	IDOutput    string `xorm:"'id_output'"`
	IDBackend   string `xorm:"'id_backend'"`
	BackendType string `xorm:"'backend_type'"`
}

// DBConfig read from DB
type DBConfig struct {
	Metrics      map[string]*SnmpMetricCfg
	Measurements map[string]*MeasurementCfg
	MFilters     map[string]*MeasFilterCfg
	GetGroups    map[string]*MGroupsCfg
	SnmpDevice   map[string]*SnmpDeviceCfg
	Influxdb     map[string]*InfluxCfg
	Kafka        map[string]*KafkaCfg
	// NEW
	Outputs map[string]*OutputCfg

	VarCatalog map[string]interface{}
}

/*
InitMetricsCfg this function does 2 things
1.- Initialice id from key of maps for all SnmpMetricCfg and InfluxMeasurementCfg objects
2.- Initialice references between InfluxMeasurementCfg and SnmpMetricGfg objects
*/
// InitMetricsCfg xx
func InitMetricsCfg(cfg *DBConfig) error {
	// TODO:
	// - check duplicates OID's => warning messages
	// Initialize references to SnmpMetricGfg into InfluxMeasurementCfg
	log.Debug("--------------------Initializing Config metrics-------------------")
	log.Debug("Initializing SNMPMetricconfig...")
	for mKey, mVal := range cfg.Metrics {
		err := mVal.Init()
		if err != nil {
			log.Warnln("Error in Metric config:", err)
			// if some error int the format the metric is deleted from the config
			delete(cfg.Metrics, mKey)
		}
	}
	log.Debug("Initializing MEASSUREMENTSconfig...")
	for mKey, mVal := range cfg.Measurements {
		err := mVal.Init(&cfg.Metrics, cfg.VarCatalog)
		if err != nil {
			log.Warnln("Error in Measurement config:", err)
			// if some error int the format the metric is deleted from the config
			delete(cfg.Metrics, mKey)
		}

		log.Debugf("FIELDMETRICS: %+v", mVal.FieldMetric)
	}
	log.Debug("-----------------------END Config metrics----------------------")
	return nil
}

// var DBConfig SQLConfig
