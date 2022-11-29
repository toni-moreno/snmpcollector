package backend

import "github.com/influxdata/telegraf"

// Declare DummyBackend
type DummyBackend struct {
	ID string
}

func NewNotInitDummyDB() *DummyBackend {
	return &DummyBackend{ID: "dummy"}
}

func (db *DummyBackend) Write([]telegraf.Metric) error {
	return nil
}

func (db *DummyBackend) Connect() error {
	return nil
}

func (db *DummyBackend) Close() error {
	return nil
}
