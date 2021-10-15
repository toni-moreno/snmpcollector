package measurement

// LOG for Measurement Objects

// Debugf info
func (m *Measurement) Debugf(expr string, vars ...interface{}) {
	expr2 := "MEASUREMENT [" + m.cfg.ID + "] " + expr
	m.Log.Debugf(expr2, vars...)
}

// Debug info
func (m *Measurement) Debug(expr string) {
	expr2 := "MEASUREMENT [" + m.cfg.ID + "] " + expr
	m.Log.Debug(expr2)
}

// Infof info
func (m *Measurement) Infof(expr string, vars ...interface{}) {
	expr2 := "MEASUREMENT [" + m.cfg.ID + "] " + expr
	m.Log.Infof(expr2, vars...)
}

// Errorf info
func (m *Measurement) Errorf(expr string, vars ...interface{}) {
	expr2 := "MEASUREMENT [" + m.cfg.ID + "] " + expr
	m.Log.Errorf(expr2, vars...)
}

// Warnf log warn data
func (m *Measurement) Warnf(expr string, vars ...interface{}) {
	expr2 := "MEASUREMENT [" + m.cfg.ID + "] " + expr
	m.Log.Warnf(expr2, vars...)
}
