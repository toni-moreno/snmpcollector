package output

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	models "github.com/influxdata/telegraf/models"
	"github.com/sirupsen/logrus"
	"github.com/toni-moreno/snmpcollector/pkg/config"
)

var (
	log *logrus.Logger
)

// SetLogger adds a logger to this module
func SetLogger(l *logrus.Logger) {
	log = l
}

// Output interface defines the required methods to be considered as an valid snmpcollector output
type Backend interface {
	Write([]telegraf.Metric) error
	Connect() error
	Close() error
}

type SinkDB struct {
	Cfg     *config.OutputCfg
	stats   SinkDBStats
	tick    *time.Ticker
	chExit  chan bool
	buffer  *models.Buffer
	Backend Backend
	// statuses
	initialized bool
	started     bool
	connected   bool
	// need to define special closed status due to several calls on exit
	// produces a panic on already closed channels (internal backend logic)
	closed  bool
	imutex  sync.Mutex
	smutex  sync.Mutex
	cmutex  sync.Mutex
	ccmutex sync.Mutex
	// special dummy state...
	dummy bool
}

// GetResetStats return outdb stats and reset its counters
func (db *SinkDB) GetResetStats() *SinkDBStats {
	if db.dummy {
		return &SinkDBStats{}
	}
	log.Infof("%s [%s-%s] - Stats: Reseting SinkDBStats for DB %s", db.Cfg.ID, db.Cfg.Backend, db.Cfg.BackendType, db.Cfg.ID)
	stt := db.stats.GetResetStats()
	log.Infof("%s [%s-%s] - Stats: stats added to buffer: %+v", db.Cfg.ID, db.Cfg.Backend, db.Cfg.BackendType, stt)
	return stt
}

// Init initializes the needed SinkDB resources and tries the backend connection
// it registers the different sinkdb statuses:
// - initialized - init has been called, no more inits should be used, even if the connection is unsuccessfull
// - started - the sender is being called and threaded within a goroutine, as it is called from external funcionts
// - connected - the output has been succesfully connected and it is registered to be able to write metrics
func (db *SinkDB) Init() error {

	if db.Backend == nil {
		return fmt.Errorf("no output found")
	}
	// enforces the initialization to guarantee that init is called once
	chset := db.CheckAndSetInitialized()
	if chset {
		log.Infof("%s [%s-%s] - Initialized - Skipping, sender thread already initialized", db.Cfg.ID, db.Cfg.Backend, db.Cfg.BackendType)
		return nil
	}
	// create only once the buffer, buffer registers internal stats based on the touple - id + backend name
	// the ID of the SinkDB config is unique, so it is guaranteed
	db.buffer = models.NewBuffer(db.Cfg.ID, db.Cfg.Backend, db.Cfg.BufferSize)

	// create needed channels and initialize the ticker
	// channels to interact with the go sender
	db.chExit = make(chan bool)
	if db.Cfg.FlushInterval == 0 {
		log.Warnf("retrieved a flushinterval=0, forcing to 30s")
		db.Cfg.FlushInterval = 30
	}
	db.tick = time.NewTicker(time.Duration(db.Cfg.FlushInterval) * time.Second)

	// tries to connect based on generic method and mark the connected status based on the result
	err := db.Backend.Connect()
	if err != nil {
		log.Errorf("Unable to connect with backend %s - %s", db.Cfg.Backend, db.Cfg.BackendType)
		db.CheckAndUnSetConnected()
		// the retrieved error should be passed to top, as we will try to reconnect during the flush interval ticker
		return nil
	}
	db.CheckAndSetConnected()
	return nil
}

// End release DB connection
func (db *SinkDB) End() {
	if db.CheckAndUnSetInitialized() {
		close(db.chExit)
		db.tick.Stop()
		if !db.IsClosed() {
			err := db.Backend.Close()
			if err != nil {
				log.Errorf("unable to close output %s, error: %s", db.Cfg.ID, err)
			}
			db.CheckAndSetClosed()
		}
	}
}

// CheckAndSetConnected forces the backend status to connected and returns the last status
func (db *SinkDB) CheckAndSetConnected() bool {
	log.Debugf("Connected set as true")
	db.cmutex.Lock()
	defer db.cmutex.Unlock()
	retval := db.connected
	db.connected = true
	return retval
}

// CheckAndUnSetConnected force the backend status to unconnected and returns the last status
func (db *SinkDB) CheckAndUnSetConnected() bool {
	db.cmutex.Lock()
	defer db.cmutex.Unlock()
	retval := db.connected
	db.connected = false
	return retval
}

// IsConnected check if the backend backend is already connected
func (db *SinkDB) IsConnected() bool {
	db.cmutex.Lock()
	defer db.cmutex.Unlock()
	return db.connected
}

// CheckAndSetInitialized forces the backend status to initialized and returns the last status
func (db *SinkDB) CheckAndSetInitialized() bool {
	db.imutex.Lock()
	defer db.imutex.Unlock()
	retval := db.initialized
	db.initialized = true
	return retval
}

// CheckAndUnSetInitialized forces the backend status to non-initialized and returns the last status
func (db *SinkDB) CheckAndUnSetInitialized() bool {
	db.imutex.Lock()
	defer db.imutex.Unlock()
	retval := db.initialized
	db.initialized = false
	return retval
}

// IsInitialized check if the backend backend is already initialized
func (db *SinkDB) IsInitialized() bool {
	db.imutex.Lock()
	defer db.imutex.Unlock()
	return db.initialized
}

// CheckAndSetInitialized forces the backend status to started and returns the last status
func (db *SinkDB) CheckAndSetStarted() bool {
	db.smutex.Lock()
	defer db.smutex.Unlock()
	retval := db.started
	db.started = true
	return retval
}

// CheckAndSetInitialized forces the backend status to not-started and returns the last status
func (db *SinkDB) CheckAndUnSetStarted() bool {
	db.smutex.Lock()
	defer db.smutex.Unlock()
	retval := db.started
	db.started = false
	return retval
}

// IsStarted check if the backend backend is already started
func (db *SinkDB) IsStarted() bool {
	db.smutex.Lock()
	defer db.smutex.Unlock()
	return db.started
}

// CheckAndSetInitialized forces the backend status to started and returns the last status
func (db *SinkDB) CheckAndSetClosed() bool {
	db.ccmutex.Lock()
	defer db.ccmutex.Unlock()
	retval := db.closed
	db.closed = true
	return retval
}

// CheckAndSetInitialized forces the backend status to not-started and returns the last status
func (db *SinkDB) CheckAndUnSetClosed() bool {
	db.ccmutex.Lock()
	defer db.ccmutex.Unlock()
	retval := db.closed
	db.closed = false
	return retval
}

// IsStarted check if the backend backend is already started
func (db *SinkDB) IsClosed() bool {
	db.ccmutex.Lock()
	defer db.ccmutex.Unlock()
	return db.started
}

// StopSender finalize sender goroutines
func (db *SinkDB) StopSender() {
	if db.IsStarted() {
		db.chExit <- true
		return
	}
	log.Infof("Cannot stop sender [%s] - it's already stopped", db.Cfg.ID)
}

// SendToBuffer retrieves the telegraf metrics fom the measurements or selfmon to add them into the buffer
func (db *SinkDB) SendToBuffer(metrics []telegraf.Metric, origin string) (int, error) {
	// check for dummy and skip buffer send
	if db.Cfg.BackendType == "dummy" {
		return 0, nil
	}
	// check for enqueueonerror and active
	// if active = false and enqueueonerror = false, we should not add metrics to the buffer
	// metrics are considered as discarded
	if !db.Cfg.Active && !db.Cfg.EnqueueOnError {
		log.Warnf("%s - Skipped metrics, SinkDB is disabled and EnqueueOnError is set to false", origin)
		return len(metrics), nil
	}
	log.Debugf("Sending %d metrics from [%s] to the buffer", len(metrics), origin)
	dropped := db.buffer.Add(metrics...)
	if dropped > 0 {
		// need to replicate the log to appear on stdout and on specific device log
		log.Errorf("unable to add metrics to the buffer, dropped %d of total %d metrics from %s", dropped, len(metrics), origin)
		return dropped, fmt.Errorf("unable to add metrics to the buffer, dropped %d metrics of total %d from %s", dropped, len(metrics), origin)
	}

	return len(metrics), nil
}

// StartSender starts the sender goroutine only once, this method is called on each device initialization
func (db *SinkDB) StartSender(wg *sync.WaitGroup) {
	if db.CheckAndSetStarted() {
		log.Infof("%s [%s-%s] - Started - Skipping, sender thread already started", db.Cfg.ID, db.Cfg.Backend, db.Cfg.BackendType)
		return
	}
	wg.Add(1)
	go db.startSenderGo(rand.Int(), wg)
}

func (db *SinkDB) startSenderGo(r int, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Infof("%s [%s-%s] - Beginning Sender thread", db.Cfg.ID, db.Cfg.Backend, db.Cfg.BackendType)

	// start the loop
	for {
		select {
		case <-db.chExit:
			if !db.IsConnected() {
				log.Infof("Output - %s is not connected, try to connect again", db.Cfg.ID)
				if err := db.Backend.Connect(); err == nil {
					db.CheckAndSetConnected()
					log.Infof("Output - %s successfully connected!", db.Cfg.ID)
				} else {
					log.Errorf("Output - %s failed trying to connect on this iteration, exitting without flushing data: %s", db.Cfg.ID, err)
					return
				}
			}
			// TODO: pending to force a full flush without batches
			// single batch with full size?
			t := time.Now()
			nBuffer := db.buffer.Len()
			nBatches := nBuffer/db.Cfg.BufferSize + 1
			for i := 0; i < nBatches; i++ {
				mett := db.buffer.Batch(db.Cfg.MetricBatchSize)
				log.Debugf("%s [%s-%s] - Selected Batchsize of - %d [%s]", db.Cfg.ID, db.Cfg.Backend, db.Cfg.BackendType, db.buffer.Len(), t)
				// avoid to write 0 metrics
				if len(mett) == 0 {
					break
				}
				err := db.Backend.Write(mett)
				if err != nil {
					// if error happens, we should discard the metrics and remove them, forcing and accept
					log.Errorf("%s [%s-%s] Error trying to write metrics %d / %d, dropped, error: %s", db.Cfg.ID, db.Cfg.Backend, db.Cfg.BackendType, len(mett), db.buffer.Len(), err)
				} else {
					log.Infof("%s [%s-%s] Succesfull write metrics %d / %d", db.Cfg.ID, db.Cfg.Backend, db.Cfg.BackendType, len(mett), db.buffer.Len())
				}
				db.buffer.Accept(mett)
			}
			//need to flush all data
			log.Infof("EXIT from Influx sender process for device [%s] ", db.Cfg.ID)
			return
		case t := <-db.tick.C:
			// declare the status needed vars
			var ps, pd, fs, we, ws int64
			var bl float32
			var wt time.Duration

			// as the buffer is entirely flushed we need to retrieve the buffer length prior to the write process
			// the bufferlength stat should be computed as the max of all ticks on a gather freq
			bl = float32(db.buffer.Len()*100.0) / float32(db.Cfg.BufferSize)

			// if db is not active we should not try to send/connect to the output
			if !db.Cfg.Active {
				log.Debugf("Output - %s with backend %s is disabled, not sending metrics", db.Cfg.ID, db.Cfg.Backend)
				// as the buffersize can change due to enqueueonerror, we should update as if they are reseted
				db.stats.FillStats(0, 0, 0, 0, 0, 0, bl)
				continue
			}

			// we check if the output is already connected, if it isn't, try it again
			// metrics keep being stored on
			if !db.IsConnected() {
				log.Infof("Output - %s with backend %s is not connected, try to connect again", db.Cfg.ID, db.Cfg.Backend)
				if err := db.Backend.Connect(); err == nil {
					db.CheckAndSetConnected()
					log.Infof("Output - %s successfully connected!", db.Cfg.ID)
				} else {
					log.Errorf("Output - %s failed trying to connect on this iteration, error: %s", db.Cfg.ID, err)
					continue
				}
			}
			// TODO: apply jitter, just before the write process
			// flush_jitter
			// end jitter
			nBuffer := db.buffer.Len()
			nBatches := nBuffer/db.Cfg.MetricBatchSize + 1
			for i := 0; i < nBatches; i++ {
				mett := db.buffer.Batch(db.Cfg.MetricBatchSize)
				log.Debugf("%s [%s-%s] - Selected Batchsize of - %d [%s]", db.Cfg.ID, db.Cfg.Backend, db.Cfg.BackendType, db.buffer.Len(), t)
				// avoid to write 0 metrics
				if len(mett) == 0 {
					break
				}
				st := time.Now()
				err := db.Backend.Write(mett)
				wt += time.Since(st)
				if err != nil {
					// if error happens, we should reject those metrics and they are inserted again in the buffer
					if db.Cfg.EnqueueOnError {
						db.buffer.Reject(mett)
						log.Warnf("%s [%s-%s] Error trying to write metrics %d / %d, enqueud, error: %s", db.Cfg.ID, db.Cfg.Backend, db.Cfg.BackendType, len(mett), db.buffer.Len(), err)
					} else {
						db.buffer.Accept(mett)
						pd += int64(len(mett))
						log.Warnf("%s [%s-%s] Error trying to write metrics %d / %d, dropped, error: %s", db.Cfg.ID, db.Cfg.Backend, db.Cfg.BackendType, len(mett), db.buffer.Len(), err)
					}
					we++
				} else {
					log.Debugf("%s [%s-%s] Succesfull write metrics %d / %d", db.Cfg.ID, db.Cfg.Backend, db.Cfg.BackendType, len(mett), db.buffer.Len())
					db.buffer.Accept(mett)
					ws++
					ps += int64(len(mett))
					for _, met := range mett {
						fs += int64(len(met.FieldList()))
					}
				}
			}
			// batches are finished, we update the stats with the gathered
			log.Debugf("%s [%s-%s] - After %d batches - Stats: PointSent: %d, PointsDropped: %d, FieldSent: %d, WriteSent %d, WriteSentError %d, BufferLen: %f, WriteTime: %v",
				db.Cfg.ID, db.Cfg.Backend, db.Cfg.BackendType, nBatches, ps, pd, fs, ws, we, bl, wt)
			db.stats.FillStats(ps, pd, fs, ws, we, wt, bl)
		}
	}
}
