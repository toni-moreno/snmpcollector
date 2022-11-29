package output

import (
	"sync"
	"time"
)

// SinkDBStats get stats
type SinkDBStats struct {
	// Fields Sent
	FieldSent int64
	// Field Sent the max
	FieldSentMax int64
	// Points Sent
	PSent int64
	// PSentMax the max
	PSentMax int64
	// Points dropped
	PDropped int64
	// WriteSent BatchPoints sent
	WriteSent int64
	// WriteErrors BatchPoints with errors
	WriteErrors int64
	// WriteTime
	WriteTime time.Duration
	// WriteTimeMax
	WriteTimeMax time.Duration
	// BufferPercentUsed
	BufferPercentUsed float32
	mutex             sync.Mutex
}

// GetResetStats get stats for this InfluxStats Output
func (is *SinkDBStats) GetResetStats() *SinkDBStats {
	is.mutex.Lock()
	defer is.mutex.Unlock()
	retstat := &SinkDBStats{
		FieldSent:         is.FieldSent,
		FieldSentMax:      is.FieldSentMax,
		PSent:             is.PSent,
		PSentMax:          is.PSentMax,
		PDropped:          is.PDropped,
		WriteSent:         is.WriteSent,
		WriteErrors:       is.WriteErrors,
		WriteTime:         is.WriteTime,
		WriteTimeMax:      is.WriteTimeMax,
		BufferPercentUsed: is.BufferPercentUsed,
	}
	is.FieldSent = 0
	is.FieldSentMax = 0
	is.PSent = 0
	is.PDropped = 0
	is.PSentMax = 0
	is.WriteSent = 0
	is.WriteErrors = 0
	is.WriteTime = 0
	is.WriteTimeMax = 0
	is.BufferPercentUsed = 0
	return retstat
}

// FillStats updates in threadsafe the current SinkDBStats
func (is *SinkDBStats) FillStats(ps, pd, fs, ws, we int64, wt time.Duration, bpUsed float32) {
	is.mutex.Lock()
	defer is.mutex.Unlock()
	// PointSent
	if is.PSentMax < ps {
		is.PSentMax = ps
	}
	is.PSent += ps

	// PDropped
	is.PDropped += pd

	// Write Time
	if is.WriteTimeMax < wt {
		is.WriteTimeMax = wt
	}
	is.WriteTime += wt

	// FieldSent
	if is.FieldSentMax < fs {
		is.FieldSentMax = fs
	}
	is.FieldSent += fs

	// WriteSent, WriteErrors
	is.WriteSent += ws
	is.WriteErrors += we

	// BufferPercentUsed
	if is.BufferPercentUsed < bpUsed {
		is.BufferPercentUsed = bpUsed
	}
}
