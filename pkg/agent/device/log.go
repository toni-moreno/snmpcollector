package device

// Debugf info
func (d *SnmpDevice) Debugf(expr string, vars ...interface{}) {
	expr2 := "SNMPDEVICE [" + d.cfg.ID + "] " + expr
	d.log.Debugf(expr2, vars...)
}

// Infof info
func (d *SnmpDevice) Infof(expr string, vars ...interface{}) {
	expr2 := "SNMPDEVICE [" + d.cfg.ID + "] " + expr
	d.log.Infof(expr2, vars...)
}

// Errorf info
func (d *SnmpDevice) Errorf(expr string, vars ...interface{}) {
	expr2 := "SNMPDEVICE [" + d.cfg.ID + "] " + expr
	d.log.Errorf(expr2, vars...)
}

// Debugf info
func (d *SnmpDevice) Warnf(expr string, vars ...interface{}) {
	expr2 := "SNMPDEVICE [" + d.cfg.ID + "] " + expr
	d.log.Warnf(expr2, vars...)
}
