package measurement

// Debugf info
func (m *Measurement) Debugf(expr string, vars ...interface{}) {
	expr2 := "MEASUREMENT [" + m.cfg.ID + "] " + expr
	m.log.Debugf(expr2, vars...)
}

// Debug info
func (m *Measurement) Debug(expr string) {
	expr2 := "MEASUREMENT [" + m.cfg.ID + "] " + expr
	m.log.Debug(expr2)
}

// Infof info
func (m *Measurement) Infof(expr string, vars ...interface{}) {
	expr2 := "MEASUREMENT [" + m.cfg.ID + "] " + expr
	m.log.Infof(expr2, vars...)
}

// Errorf info
func (m *Measurement) Errorf(expr string, vars ...interface{}) {
	expr2 := "MEASUREMENT [" + m.cfg.ID + "] " + expr
	m.log.Errorf(expr2, vars...)
}

// Debugf info
func (m *Measurement) Warnf(expr string, vars ...interface{}) {
	expr2 := "MEASUREMENT [" + m.cfg.ID + "] " + expr
	m.log.Warnf(expr2, vars...)
}
