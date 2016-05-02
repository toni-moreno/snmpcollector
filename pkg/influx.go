package main

import (
	"fmt"
	//"log"
	"strings"
	"time"
	//"sync"
	"sync/atomic"

	"github.com/influxdata/influxdb/client/v2"
)

type InfluxConfig struct {
	Host      string `toml:"host"`
	Port      int    `toml:"port"`
	DB        string `toml:"db"`
	User      string `toml:"user"`
	Password  string `toml:"password"`
	Retention string `toml:"retention"`
	iChan     chan *client.BatchPoints
	client    client.Client
	Sent      int64
	Errors    int64
}

func (c *InfluxConfig) incSent() {
	atomic.AddInt64(&c.Sent, 1)
}

func (c *InfluxConfig) addSent(n int64) {
	atomic.AddInt64(&c.Sent, n)
}

func (c *InfluxConfig) incErrors() {
	atomic.AddInt64(&c.Errors, 1)
}

func (c *InfluxConfig) addErrors(n int64) {
	atomic.AddInt64(&c.Errors, n)
}

func (c *InfluxConfig) BP() *client.BatchPoints {
	if len(c.Retention) == 0 {
		c.Retention = "default"
	}
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:        c.DB,
		RetentionPolicy: c.Retention,
		Precision:       "ns", //Default precision for Time lib
	})
	return &bp
}

func (c *InfluxConfig) Connect() error {
	conf := client.HTTPConfig{
		Addr:     fmt.Sprintf("http://%s:%d", c.Host, c.Port),
		Username: c.User,
		Password: c.Password,
	}
	cli, err := client.NewHTTPClient(conf)
	c.client = cli
	if err != nil {
		return err
	}

	_, _, err = c.client.Ping(time.Duration(5))
	return err
}

func (c *InfluxConfig) Init() {

	if verbose {
		log.Infoln("Connecting to: ", c.Host)
	}
	c.iChan = make(chan *client.BatchPoints, 65535)
	if err := c.Connect(); err != nil {
		log.Errorln("failed connecting to: ", c.Host)
		log.Errorln("error: ", err)
		log.Fatal(err)
	}
	if verbose {
		log.Infoln("Connected to: ", c.Host)
	}

	go influxEmitter(c)
}

func (c *InfluxConfig) Send(bps *client.BatchPoints) {
	c.iChan <- bps
}

func (c *InfluxConfig) Hostname() string {
	return strings.Split(c.Host, ":")[0]
}

// use chan as a queue so that interupted connections to
// influxdb server don't drop collected data

func influxEmitter(c *InfluxConfig) {
	log.Info("beggining Influx Emmiter thread")
	for {
		select {
		case data := <-c.iChan:
			/*if testing {
				break
			}*/
			if data == nil {
				log.Warn("null influx input")
				continue
			}

			// keep trying until we get it (don't drop the data)
			for {
				if err := c.client.Write(*data); err != nil {
					c.incErrors()
					log.Errorln("influxdb write error: ", err)
					// try again in a bit
					// TODO: this could be better
					// Todo add InfluxResend on error.
					time.Sleep(30 * time.Second)
					continue
				} else {
					c.incSent()
				}
				break
			}
		}
	}
}
