package device

import ()

// DevStat minimal info to show users
type DevStat struct {
	Requests           int64
	Gets               int64
	Errors             int64
	ReloadLoopsPending int
	DeviceActive       bool
	DeviceConnected    bool
	NumMeasurements    int
	NumMetrics         int
}

// GetBasicStats get basic info for this device
func (d *SnmpDevice) GetBasicStats() *DevStat {

	sum := 0
	for _, m := range d.Measurements {
		sum += len(m.OidSnmpMap)
	}
	d.mutex.Lock()
	stat := &DevStat{
		Requests:           d.Requests,
		Gets:               d.Gets,
		Errors:             d.Errors,
		ReloadLoopsPending: d.ReloadLoopsPending,
		DeviceActive:       d.DeviceActive,
		DeviceConnected:    d.DeviceConnected,
		NumMeasurements:    len(d.Measurements),
		NumMetrics:         sum,
	}
	d.mutex.Unlock()
	return stat
}

func (d *SnmpDevice) addRequests(n int64) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.Requests += n
}

func (d *SnmpDevice) resetCounters() {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.Gets = 0
	d.Errors = 0
}

func (d *SnmpDevice) addGets(n int64) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.Gets += n
}

func (d *SnmpDevice) addErrors(n int64) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.Errors += n
}
