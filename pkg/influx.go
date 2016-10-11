package main

import (
	"fmt"
	"github.com/influxdata/influxdb/client/v2"
	"math/rand"
	"strings"
	"sync/atomic"
	"time"
)

/*InfluxDB database export */
type InfluxDB struct {
	cfg     *InfluxCfg
	started bool
	dummy   bool
	iChan   chan *client.BatchPoints
	client  client.Client
	Sent    int64
	Errors  int64
}

var influxdbDummy = &InfluxDB{
	cfg:     nil,
	started: false,
	dummy:   true,
	iChan:   nil,
	client:  nil,
	Sent:    0,
	Errors:  0,
}

func (db *InfluxDB) incSent() {
	atomic.AddInt64(&db.Sent, 1)
}

func (db *InfluxDB) addSent(n int64) {
	atomic.AddInt64(&db.Sent, n)
}

func (db *InfluxDB) incErrors() {
	atomic.AddInt64(&db.Errors, 1)
}

func (db *InfluxDB) addErrors(n int64) {
	atomic.AddInt64(&db.Errors, n)
}

//BP create a Batch point influx object
func (db *InfluxDB) BP() *client.BatchPoints {
	if db.dummy == true {
		bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
			Database:        "dbdummy",
			RetentionPolicy: "dbretention",
			Precision:       "ns", //Default precision for Time lib
		})
		return &bp
	}
	//
	if len(db.cfg.Retention) == 0 {
		db.cfg.Retention = "autogen"
	}
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:        db.cfg.DB,
		RetentionPolicy: db.cfg.Retention,
		Precision:       "ns", //Default precision for Time lib
	})
	return &bp
}

//Connect to influxdb
func (db *InfluxDB) Connect() error {
	if db.dummy == true {
		return nil
	}
	conf := client.HTTPConfig{
		Addr:     fmt.Sprintf("http://%s:%d", db.cfg.Host, db.cfg.Port),
		Username: db.cfg.User,
		Password: db.cfg.Password,
	}
	cli, err := client.NewHTTPClient(conf)
	db.client = cli
	if err != nil {
		return err
	}

	_, _, err = db.client.Ping(time.Duration(5))
	return err
}

//Init initialies runtime info
func (db *InfluxDB) Init() {
	if db.dummy == true {
		return
	}
	if db.started == true {
		log.Infof("Emitter thread to : %s  already started (skipping Initialization)", db.cfg.ID)
		return
	}

	log.Infof("Initializing influxdb with id = %s", db.cfg.ID)

	if verbose {
		log.Infoln("Connecting to: ", db.cfg.Host)
	}
	db.iChan = make(chan *client.BatchPoints, 65535)
	if err := db.Connect(); err != nil {
		log.Errorln("failed connecting to: ", db.cfg.Host)
		log.Errorln("error: ", err)
		log.Fatal(err)
	}
	if verbose {
		log.Infoln("Connected to: ", db.cfg.Host)
	}
	db.started = true
	go influxEmitter(db, rand.Int())
}

//Send send data
func (db *InfluxDB) Send(bps *client.BatchPoints) {
	if db.dummy == true {
		return
	}
	db.iChan <- bps
}

//Hostname get hostname
func (db *InfluxDB) Hostname() string {
	return strings.Split(db.cfg.Host, ":")[0]
}

// use chan as a queue so that interupted connections to
// influxdb server don't drop collected data

func influxEmitter(db *InfluxDB, r int) {
	log.Infof("beggining Influx Emmiter thread: %d", r)
	for {
		select {
		case data := <-db.iChan:
			/*if testing {
				break
			}*/
			if data == nil {
				log.Warn("null influx input")
				continue
			}

			// keep trying until we get it (don't drop the data)
			log.Debugf("sending data from Emmiter %d", r)
			for {
				if err := db.client.Write(*data); err != nil {
					db.incErrors()
					log.Errorln("influxdb write error: ", err)
					// try again in a bit
					// TODO: this could be better
					// Todo add InfluxResend on error.
					time.Sleep(30 * time.Second)
					continue
				} else {
					db.incSent()
				}
				break
			}
		}
	}
}
