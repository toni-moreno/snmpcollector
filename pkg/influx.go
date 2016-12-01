package main

import (
	"fmt"
	"github.com/influxdata/influxdb/client/v2"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

/*InfluxDB database export */
type InfluxDB struct {
	cfg     *InfluxCfg
	started bool
	dummy   bool
	iChan   chan *client.BatchPoints
	chExit  chan bool
	client  client.Client
	Sent    int64
	Errors  int64
	mutex   sync.Mutex
}

var influxdbDummy = &InfluxDB{
	cfg:     nil,
	started: false,
	dummy:   true,
	iChan:   nil,
	chExit:  nil,
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
		Addr:      fmt.Sprintf("http://%s:%d", db.cfg.Host, db.cfg.Port),
		Username:  db.cfg.User,
		Password:  db.cfg.Password,
		UserAgent: db.cfg.UserAgent,
		Timeout:   time.Duration(db.cfg.Timeout) * time.Second,
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
	db.mutex.Lock()
	started := db.started
	db.mutex.Unlock()
	if started == true {
		log.Infof("Sender thread to : %s  already Initialized (skipping Initialization)", db.cfg.ID)
		return
	}

	if len(db.cfg.UserAgent) == 0 {
		db.cfg.UserAgent = "snmpCollector-" + db.cfg.ID
	}

	log.Infof("Initializing influxdb with id = %s", db.cfg.ID)

	log.Infof("Connecting to: %s", db.cfg.Host)
	db.iChan = make(chan *client.BatchPoints, 65535)
	db.chExit = make(chan bool)
	if err := db.Connect(); err != nil {
		log.Errorln("failed connecting to: ", db.cfg.Host)
		log.Errorln("error: ", err)
		//if no connection done started = false and it will try to test again later??
		return
	}

	log.Infof("Connected to: %s", db.cfg.Host)
}

// End finalize sender goroutines
func (db *InfluxDB) StopSender() {
	if db.dummy == true {
		return
	}
	db.mutex.Lock()
	started := db.started
	db.mutex.Unlock()
	if started == true {
		db.chExit <- true
	}

	log.Info("Can not stop Sender %s becaouse of its already stopped", db.cfg.ID)
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

// StartSender begins sender loop
func (db *InfluxDB) StartSender(wg *sync.WaitGroup) {
	db.mutex.Lock()
	started := db.started
	db.mutex.Unlock()
	if started == true {
		log.Infof("Sender thread to : %s  already started (skipping Initialization)", db.cfg.ID)
		return
	}

	go db.startSenderGo(rand.Int(), wg)
}

func (db *InfluxDB) startSenderGo(r int, wg *sync.WaitGroup) {
	defer wg.Done()
	wg.Add(1)
	db.mutex.Lock()
	db.started = true
	db.mutex.Unlock()
	log.Infof("beggining Influx Sender thread: %d", r)
	for {
		select {
		case <-db.chExit:
			log.Infof("EXIT from Influx sender process for device %s ", db.cfg.ID)
			db.mutex.Lock()
			db.started = false
			db.mutex.Unlock()
			return
		case data := <-db.iChan:
			if data == nil {
				log.Warn("null influx input")
				continue
			}

			// keep trying until we get it (don't drop the data)
			log.Debugf("sending data from Sender %d", r)
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
