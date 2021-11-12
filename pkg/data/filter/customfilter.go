package filter

import (
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"github.com/toni-moreno/snmpcollector/pkg/data/utils"
)

// CustomFilter a filter created from a device query, usuali used only with the DeviceOrigin , but could be used on more than one but
type CustomFilter struct {
	filterLabels map[string]string
	CustomID     string
	EnableAlias  bool
	dbc          *config.DatabaseCfg

	log utils.Logger
}

// NewCustomFilter creates a new CustomFilter
func NewCustomFilter(cid string, enableAlias bool, l utils.Logger) *CustomFilter {
	return &CustomFilter{CustomID: cid, EnableAlias: enableAlias, log: l}
}

// Init Load Confiration before use them
func (cf *CustomFilter) Init(arg ...interface{}) error {
	cf.dbc = arg[0].(*config.DatabaseCfg)
	cf.log.Infof("CUSTOMFILTER [%s] Init CustomFilter Enable Alias: %t", cf.CustomID, cf.EnableAlias)
	_, err := cf.dbc.GetCustomFilterCfgByID(cf.CustomID)
	if err != nil {
		cf.log.Debugf("CUSTOMFILTER [%s] Error on define custom Filter %s", cf.CustomID, err)
		return err
	}

	cf.filterLabels = make(map[string]string)
	return nil
}

// Count Get current items in the filter
func (cf *CustomFilter) Count() int {
	return len(cf.filterLabels)
}

// MapLabels return the final tagmap from all posible values and the filter results
func (cf *CustomFilter) MapLabels(AllIndexedLabels map[string]string) map[string]string {
	curIndexedLabels := make(map[string]string, len(cf.filterLabels))
	for kf, vf := range cf.filterLabels {
		for kl, vl := range AllIndexedLabels {
			if kf == vl {
				if len(vf) > 0 {
					// map[kl]vf (alias to key of the label
					curIndexedLabels[kl] = vf
				} else {
					// map[kl]vl (original name)
					curIndexedLabels[kl] = vl
				}
			}
		}
	}
	return curIndexedLabels
}

// Update load filtered data from Database config online time
func (cf *CustomFilter) Update() error {
	cf.log.Infof("CUSTOMFILTER [%s] apply CustomFilter Enable Alias: %t", cf.CustomID, cf.EnableAlias)
	// reset current filter
	cf.filterLabels = make(map[string]string)

	filter, err := cf.dbc.GetCustomFilterCfgByID(cf.CustomID)
	cf.log.Debugf("CUSTOMFILTER [%s] custom Filter %+v", cf.CustomID, filter)

	if err != nil {
		return err
	}

	for _, vf := range filter.Items {
		if cf.EnableAlias {
			cf.filterLabels[vf.TagID] = vf.Alias
		} else {
			cf.filterLabels[vf.TagID] = ""
		}
	}
	return nil
}
