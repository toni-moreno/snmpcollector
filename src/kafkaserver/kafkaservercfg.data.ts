export const KafkaServerCfgComponentConfig: any =
  {
    'name' : 'Outputs > Kafka Server',
    'table-columns' : [
      { title: 'ID', name: 'ID' },
      { title: 'Brokers', name: 'Brokers' },
      { title: 'Enable TLS',name:'EnableTLS'},
      { title: 'Enable Socks5',name:'Socks5ProxyEnabled'},
      { title: 'Topic', name: 'Topic' },
      { title: 'Method', name: 'Method' },
      { title: 'RoutingKey', name: 'RoutingKey' },
      { title: 'MaxRetry', name:'MaxRetry'}



      // type KafkaCfg struct {
      //   ID        string   `xorm:"'id' unique" binding:"Required"`
      //   Brokers   []string `xorm:"brokers" binding:"Required"`
      //   Topic     string   `xorm:"topic" binding:"Required;IntegerNotZero"`
      //   Method    string   `xorm:"topic_suffix_method"`
      //   Keys      []string `xorm:"topic_suffix_keys"`
      //   Separator string   `xorm:"topic_suffix_separator"`
      //   //	RoutingTag         string      `xorm:"routing_tag" binding:"Required"`
      //   RoutingKey         string `xorm:"routing_key" binding:"Required"`
      //   Description        string `xorm:"description"`
      //   RequiredAcks       int    `xorm:"'required_acks' default -1"`
      //   MaxRetry           int    `xorm:"'max_retry' default 3"`
      //   MaxMessageBytes    int    `xorm:"max_message_bytes"`
      //   IdempotentWrites   bool   `xorm:"idempotent_writes"`
      //   TLSCA              string `xorm:"tls_ca"`
      //   TLSCert            string `xorm:"tls_cert"`
      //   TLSKey             string `xorm:"tls_key"`
      //   TLSKeyPwd          string `xorm:"tls_key_pwd"`
      //   TLSMinVersion      string `xorm:"tls_min_version"`
      //   InsecureSkipVerify bool   `xorm:"insecure_skip_verify"`
      //   ServerName         string `xorm:"tls_server_name"`
      //   Version            string `xorm:"version"`
      //   ClientID           string `xorm:"client_id"`
      //   CompressionCodec   int    `xorm:"compression_codec"`
      //   EnableTLS          *bool  `xorm:"enable_tls"`
      //   // Disable full metadata fetching
      //   MetadataFull *bool `xorm:"metadata_full"`


      
      // { title: 'Enable SSL',name:'EnableSSL'},
      // { title: 'DB', name: 'DB' },
      // { title: 'User', name: 'User' },
      // { title: 'Retention', name: 'Retention' },
      // { title: 'Precision', name: 'Precision' },
      // { title: 'Timeout', name: 'Timeout' },
      // { title: 'Buffer Size', name: 'BufferSize' },
      // { title: 'User Agent', name: 'UserAgent' }
    ],
    'slug' : 'kafkacfg'
  }; 


  // type KafkaCfg struct {
  //   ID          string      `xorm:"'id' unique" binding:"Required"`
  //   Brokers     []string    `xorm:"brokers" binding:"Required"`
  //   Topic       string      `xorm:"topic" binding:"Required;IntegerNotZero"`
  //   TopicSuffix TopicSuffix `xorm:"extends" binding:"Required"`
  //   //	RoutingTag         string      `xorm:"routing_tag" binding:"Required"`
  //   RoutingKey  string `xorm:"routing_key" binding:"Required"`
  //   SSLCA       string `xorm:"ssl_ca"`
  //   SSLCert     string `xorm:"ssl_cert"`
  //   SSLKey      string `xorm:"ssl_key"`
  //   BufferSize  int    `xorm:"'buffer_size' default 65535"`
  //   Description string `xorm:"description"`
  // }

  export const TableRole : string = 'fulledit';
  export const OverrideRoleActions : Array<Object> = [
    {'name':'export', 'type':'icon', 'icon' : 'glyphicon glyphicon-download-alt text-default', 'tooltip': 'Export item'},
    {'name':'view', 'type':'icon', 'icon' : 'glyphicon glyphicon-eye-open text-success', 'tooltip': 'View item'},
    {'name':'edit', 'type':'icon', 'icon' : 'glyphicon glyphicon-edit text-warning', 'tooltip': 'Edit item'},
    {'name':'remove', 'type':'icon', 'icon' : 'glyphicon glyphicon glyphicon-remove text-danger', 'tooltip': 'Remove item'}
  ]