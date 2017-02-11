package filter

import ()

// Filter Interface to operate with filters
type Filter interface {
	Init(arg ...interface{}) error
	Update() error
	//construct the final index array from all index and filters
	MapLabels(map[string]string) map[string]string
	Count() int
}
