package backend

import (
	"fmt"
	"net/http"
	"time"

	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/influxdata/telegraf"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"github.com/toni-moreno/snmpcollector/pkg/data/utils"
)

/*InfluxDB database export */
type InfluxDB struct {
	cfg    *config.InfluxCfg
	dummy  bool
	client client.Client
}

// DummyDB a BD struct needed if no database configured
var DummyDB = &InfluxDB{
	cfg:   nil,
	dummy: true,
	// iChan:  nil,
	// chExit: nil,
	client: nil,
}

// BP create a Batch point influx object
func (db *InfluxDB) BP() (*client.BatchPoints, error) {
	if db.dummy {
		bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
			Database:        "dbdummy",
			RetentionPolicy: "dbretention",
			Precision:       "ns", //Default precision for Time lib
		})
		return &bp, nil
	}
	//
	if len(db.cfg.Retention) == 0 {
		db.cfg.Retention = "autogen"
	}
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:        db.cfg.DB,
		RetentionPolicy: db.cfg.Retention,
		Precision:       db.cfg.Precision, //Default precision for Time lib
	})
	if err != nil {
		log.Errorf("Error on create BatchPoint: %s", err)
		return nil, err
	}
	return &bp, err
}

// Ping InfluxDB Server
func Ping(cfg *config.InfluxCfg) (client.Client, time.Duration, string, error) {

	var conf client.HTTPConfig
	if cfg.EnableSSL {
		tls, err := utils.GetTLSConfig(cfg.SSLCert, cfg.SSLKey, cfg.SSLCA, cfg.InsecureSkipVerify)
		if err != nil {
			log.Errorf("Error on Create TLS config: %s", err)
			return nil, 0, "", err
		}
		conf = client.HTTPConfig{
			Addr:      fmt.Sprintf("https://%s:%d", cfg.Host, cfg.Port),
			Username:  cfg.User,
			Password:  cfg.Password,
			UserAgent: cfg.UserAgent,
			Timeout:   time.Duration(cfg.Timeout) * time.Second,
			TLSConfig: tls,
			Proxy:     http.ProxyFromEnvironment,
		}
	} else {

		conf = client.HTTPConfig{
			Addr:      fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port),
			Username:  cfg.User,
			Password:  cfg.Password,
			UserAgent: cfg.UserAgent,
			Timeout:   time.Duration(cfg.Timeout) * time.Second,
			Proxy:     http.ProxyFromEnvironment,
		}
	}
	cli, err := client.NewHTTPClient(conf)

	if err != nil {
		return cli, 0, "", err
	}
	elapsed, message, err := cli.Ping(time.Duration(cfg.Timeout) * time.Second)
	log.Infof("PING Influx Database %s : Elapsed ( %s ) : MSG : %s", cfg.ID, elapsed.String(), message)
	return cli, elapsed, message, err
}

// NewNotInitInfluxDB Create Object in memory but not initialized until ready connection needed
func NewNotInitInfluxDB(c *config.InfluxCfg) *InfluxDB {
	return &InfluxDB{
		cfg:   c,
		dummy: false,
	}
}

// TimeWriteRetry time wait
const TimeWriteRetry = 10

// Init initialies runtime info
func (db *InfluxDB) Connect() error {
	var err error

	if len(db.cfg.UserAgent) == 0 {
		db.cfg.UserAgent = "snmpCollector-" + db.cfg.ID
	}

	log.Infof("Initializing influxdb with id = [ %s ]. Connecting to: %s", db.cfg.ID, db.cfg.Host)
	db.client, _, _, err = Ping(db.cfg)
	if err != nil {
		log.Errorf("failed connecting to: %s - error: %s", db.cfg.Host, err)
		return err
	}

	log.Infof("successfully connected to %s", db.cfg.Host)
	return nil
}

// End release DB connection
func (db *InfluxDB) Close() error {
	err := db.client.Close()
	return err
}

func (db *InfluxDB) Write(metrics []telegraf.Metric) error {
	var ptarray []*client.Point
	// do we create a batchpoint and keep with the same process??...
	for _, dm := range metrics {
		pt, err := client.NewPoint(dm.Name(), dm.Tags(), dm.Fields(), dm.Time())
		if err != nil {
			log.Errorf("Cannot create influx point - %s", err)
		}
		ptarray = append(ptarray, pt)
	}
	// create batchpoints
	bpts, err := db.BP()
	if err != nil {
		log.Errorf("Cannot create batchpoints - %s", err)
		return err
	}
	if bpts != nil {
		(*bpts).AddPoints(ptarray)
		// send data
		err := db.WriteMetrics(bpts, true)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *InfluxDB) WriteMetrics(data *client.BatchPoints, enqueueonerror bool) error {
	//number points
	np := len((*data).Points())
	//number of total fields
	nf := 0
	for _, v := range (*data).Points() {
		fields, err := v.Fields()
		if err != nil {
			log.Debug("Some error happened when trying to get fields for point... ")
		} else {
			nf += len(fields)
		}
	}
	// keep trying until we get it (don't drop the data)
	startSend := time.Now()
	err := db.client.Write(*data)
	elapsedSend := time.Since(startSend)
	if err != nil {
		log.Errorf("ERROR on Write batchPoint in DB %s (%d points) | elapsed : %s | Error: %s ", db.cfg.ID, np, elapsedSend.String(), err)
		return err
	}
	log.Debugf("OK on Write batchPoint in DB %s (%d points) | elapsed : %s ", db.cfg.ID, np, elapsedSend.String())
	return err
}
