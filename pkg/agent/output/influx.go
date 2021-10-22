package output

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/sirupsen/logrus"
	"github.com/toni-moreno/snmpcollector/pkg/agent/bus"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"github.com/toni-moreno/snmpcollector/pkg/data/utils"
)

var (
	log *logrus.Logger
)

// SetLogger adds a logger to this module
func SetLogger(l *logrus.Logger) {
	log = l
}

/*SinkDB database export */
type SinkDB struct {
	cfg   *config.InfluxCfg
	stats InfluxStats  //Runtime Internal statistic
	Stats *InfluxStats //Public info for thread safe accessing to the data ()

	initialized bool
	imutex      sync.Mutex
	started     bool
	smutex      sync.Mutex
	statsData   sync.RWMutex

	dummy               bool
	iChan               chan *client.BatchPoints
	client              client.Client
	OutInfo             string
	PingTime            time.Duration
	Node                *bus.Node `json:"-"`
	Active              bool
	EnqueueOnWriteError bool
}

// DummyDB a BD struct needed if no database configured
var DummyDB = &SinkDB{
	cfg:         nil,
	initialized: false,
	started:     false,
	dummy:       true,
	iChan:       nil,
	client:      nil,
}

// Action  set an antion to the output sender
func (db *SinkDB) Action(action string) error {
	if db.dummy == true {
		return nil
	}

	if db.IsStarted() != true {
		return fmt.Errorf("Is not started")
	}
	switch action {
	case "active":
		db.Node.SendMsg(&bus.Message{Type: "setactive", Data: true})
	case "deactive":
		db.Node.SendMsg(&bus.Message{Type: "setactive", Data: false})
	case "enqueue":
		db.Node.SendMsg(&bus.Message{Type: "enqueue_policy_change", Data: true})
	case "not_enqueue":
		db.Node.SendMsg(&bus.Message{Type: "enqueue_policy_change", Data: false})
	case "flushbuffer":
		db.Node.SendMsg(&bus.Message{Type: "flushbuffer"})

	}
	return nil
}

// ToJSON return a JSON version of the device data
func (db *SinkDB) ToJSON() ([]byte, error) {

	db.statsData.RLock()
	defer db.statsData.RUnlock()
	result, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		log.Errorf("Error on Get JSON data from device")
		dummy := []byte{}
		return dummy, nil
	}
	return result, err
}

// GetBasicStats get basic info for this device
func (db *SinkDB) GetBasicStats() *InfluxStats {
	db.statsData.RLock()
	defer db.statsData.RUnlock()
	return db.Stats
}

// GetBasicStats get basic info for this device
func (db *SinkDB) GetSinkDBBasicStats() *SinkDB {
	db.statsData.RLock()
	defer db.statsData.RUnlock()
	return db
}

// GetBasicStats get basic info for this device
func (db *SinkDB) getBasicStats() *InfluxStats {
	stat := db.stats.ThSafeCopy()
	return stat
}

// GetResetStats return outdb stats and reset its counters
func (db *SinkDB) GetResetStats() *InfluxStats {
	db.statsData.RLock()
	defer db.statsData.RUnlock()
	if db.dummy == true {
		log.Debug("Reseting Influxstats for DUMMY DB ")
		return &InfluxStats{}
	}
	log.Debugf("Reseting Influxstats for DB %s", db.cfg.ID)
	db.Stats = db.stats.GetResetStats()
	return db.Stats
}

//BP create a Batch point influx object
func (db *SinkDB) BP() (*client.BatchPoints, error) {
	if db.dummy == true {
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

// Ping SinkDB Server
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

//Connect to influxdb
func (db *SinkDB) Connect() error {
	if db.dummy == true {
		return nil
	}
	var err error
	db.client, db.PingTime, db.OutInfo, err = Ping(db.cfg)
	return err
}

// CheckAndSetStarted check if this thread is already working and set if not
func (db *SinkDB) CheckAndSetStarted() bool {
	db.smutex.Lock()
	defer db.smutex.Unlock()
	retval := db.started
	db.started = true
	return retval
}

// CheckAndUnSetStarted check if this thread is already working and unset if not
func (db *SinkDB) CheckAndUnSetStarted() bool {
	db.smutex.Lock()
	defer db.smutex.Unlock()
	retval := db.started
	db.started = false
	return retval
}

// IsStarted check if this thread is already working
func (db *SinkDB) IsStarted() bool {
	db.smutex.Lock()
	defer db.smutex.Unlock()
	return db.started
}

// SetStartedAs change started state
func (db *SinkDB) SetStartedAs(st bool) {
	db.smutex.Lock()
	defer db.smutex.Unlock()
	db.started = st
}

// CheckAndSetInitialized check if this thread is already working and set if not
func (db *SinkDB) CheckAndSetInitialized() bool {
	db.imutex.Lock()
	defer db.imutex.Unlock()
	retval := db.initialized
	db.initialized = true
	return retval
}

// CheckAndUnSetInitialized check if this thread is already working and set if not
func (db *SinkDB) CheckAndUnSetInitialized() bool {
	db.imutex.Lock()
	defer db.imutex.Unlock()
	retval := db.initialized
	db.initialized = false
	return retval
}

// NewNotInitSinkDB Create Object in memory but not initialized until ready connection needed
func NewNotInitSinkDB(c *config.InfluxCfg) *SinkDB {
	return &SinkDB{
		cfg:     c,
		dummy:   false,
		started: false,
	}
}

//Init initialies runtime info
func (db *SinkDB) Init(b *bus.Bus) {
	if db.dummy == true {
		return
	}

	if db.CheckAndSetInitialized() == true {
		log.Infof("Sender thread to : %s  already Initialized (skipping Initialization)", db.cfg.ID)
		return
	}

	db.Node = bus.NewNode(db.cfg.ID)
	b.Join(db.Node)

	//begin to send as soon as receiving metrics
	db.Active = true

	if len(db.cfg.UserAgent) == 0 {
		db.cfg.UserAgent = "snmpCollector-" + db.cfg.ID
	}

	log.Infof("Initializing influxdb with id = [ %s ]", db.cfg.ID)

	log.Infof("Connecting to: %s", db.cfg.Host)
	db.resetBuffer(db.cfg.BufferSize)

	if err := db.Connect(); err != nil {
		log.Errorln("failed connecting to: ", db.cfg.Host)
		log.Errorln("error: ", err)
		//if no connection done started = false and it will try to test again later??
		return
	}
	log.Infof("Connected to: %s", db.cfg.Host)
}

// AttachToBus add this device to a communition bus
func (db *SinkDB) AttachToBus(b *bus.Bus) {
	db.Node = bus.NewNode(db.cfg.ID)
	b.Join(db.Node)
}

// LeaveBus add this device to a communition bus
func (db *SinkDB) LeaveBus(b *bus.Bus) {
	b.Leave(db.Node)
}

// End release DB connection
func (db *SinkDB) End() {
	if db.dummy == true {
		return
	}
	if db.CheckAndUnSetInitialized() == true {
		close(db.iChan)
		db.client.Close()
	}
}

// StopSender finalize sender goroutines
func (db *SinkDB) StopSender() {
	if db.dummy == true {
		return
	}

	if db.IsStarted() == true {
		db.Node.SendMsg(&bus.Message{Type: "exit"})
		return
	}

	log.Infof("Can not stop Sender [%s] becaouse of it is already stopped", db.cfg.ID)
}

//Send send data
func (db *SinkDB) Send(bps *client.BatchPoints) {
	if db.dummy == true {
		return
	}
	db.iChan <- bps
}

//Hostname get hostname
func (db *SinkDB) Hostname() string {
	return strings.Split(db.cfg.Host, ":")[0]
}

// StartSender begins sender loop
func (db *SinkDB) StartSender(wg *sync.WaitGroup) {
	if db.dummy == true {
		return
	}
	if db.CheckAndSetStarted() == true {
		log.Infof("Sender thread to : %s  already started (skipping Goroutine creation)", db.cfg.ID)
		return
	}
	wg.Add(1)
	go db.startSenderGo(rand.Int(), wg)
}

func (db *SinkDB) sendBatchPoint(data *client.BatchPoints, enqueueonerror bool) {
	var bufferPercent float32
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
	// Check if the output is active
	if !db.Active {
		// Enqueue on error is active, try to store data on buffer
		if enqueueonerror {
			if len(db.iChan) < db.cfg.BufferSize {
				db.iChan <- data
				bufferPercent = (float32(len(db.iChan)) * 100.0) / float32(db.cfg.BufferSize)
				log.Infof("Buffer OUTPUT [%s] : %.2f%%", db.cfg.ID, bufferPercent)
				db.stats.WriteErrUpdate(0, bufferPercent)
				time.Sleep(time.Duration(db.cfg.TimeWriteRetry) * time.Second)
			}
		} else {
			// Drop data if enqueue on error is false and output is not active
			log.Infof("Skipped point on output %s, output is not active and enqueue on error is false", db.cfg.ID)
		}
	} else {
		// keep trying until we get it (don't drop the data)
		startSend := time.Now()
		err := db.client.Write(*data)
		elapsedSend := time.Since(startSend)

		bufferPercent = (float32(len(db.iChan)) * 100.0) / float32(db.cfg.BufferSize)
		log.Infof("Buffer OUTPUT [%s] : %.2f%%", db.cfg.ID, bufferPercent)
		if err != nil {
			db.stats.WriteErrUpdate(elapsedSend, bufferPercent)
			log.Errorf("ERROR on Write batchPoint in DB %s (%d points) | elapsed : %s | Error: %s ", db.cfg.ID, np, elapsedSend.String(), err)
			// If the queue is not full we will resend after a while
			if enqueueonerror {
				log.Debug("queing data again...")
				if len(db.iChan) < db.cfg.BufferSize {
					db.iChan <- data
					time.Sleep(time.Duration(db.cfg.TimeWriteRetry) * time.Second)
				}
			}
		} else {
			log.Debugf("OK on Write batchPoint in DB %s (%d points) | elapsed : %s ", db.cfg.ID, np, elapsedSend.String())
			db.stats.WriteOkUpdate(int64(np), int64(nf), elapsedSend, bufferPercent)
		}
	}
}

func (db *SinkDB) resetBuffer(length int) {
	//PENDING: review db.iChan concurrency
	db.iChan = make(chan *client.BatchPoints, length)
}

func (db *SinkDB) flushBuffer() {
	chanlen := len(db.iChan) // get number of entries in the batchpoint channel
	log.Infof("Flushing %d batchpoints of data in OutDB %s ", chanlen, db.cfg.ID)
	for i := 0; i < chanlen; i++ {
		//flush them
		data := <-db.iChan
		//this process only will work if backend is  running ok elsewhere points will be lost
		db.sendBatchPoint(data, false)
	}
}

func (db *SinkDB) startSenderGo(r int, wg *sync.WaitGroup) {
	defer wg.Done()

	time.Sleep(5)
	db.EnqueueOnWriteError = db.cfg.EnqueueOnWriteError

	log.Infof("beginning Influx Sender thread: [%s]", db.cfg.ID)
	for {
		select {
		case data := <-db.iChan:
			if !db.Active {
				db.sendBatchPoint(data, db.EnqueueOnWriteError)
				continue
			}
			if data == nil {
				log.Warn("null influx input")
				continue
			}
			if db.client == nil {
				log.Warn("db Client not initialized yet!!!!!")
				continue
			}

			db.sendBatchPoint(data, db.EnqueueOnWriteError)
		case val := <-db.Node.Read:
			log.Infof("Received Message...%s: %+v", val.Type, val.Data)
			switch val.Type {
			case "resetbuffer":
				db.statsData.Lock()
				db.resetBuffer(db.cfg.BufferSize)
				db.statsData.Unlock()
			case "flushbuffer":
				db.statsData.Lock()
				db.flushBuffer()
				db.statsData.Unlock()
			case "setactive":
				db.statsData.Lock()
				db.Active = val.Data.(bool)
				db.statsData.Unlock()
			case "enqueue_policy_change":
				db.statsData.Lock()
				db.EnqueueOnWriteError = val.Data.(bool)
				db.statsData.Unlock()
			case "exit":
				log.Infof("invoked Force Exit")
				//need to flush all data
				db.flushBuffer()
				log.Infof("EXIT from Influx sender process for device [%s] ", db.cfg.ID)
				db.SetStartedAs(false)
				return

			}
		}
	}
}
