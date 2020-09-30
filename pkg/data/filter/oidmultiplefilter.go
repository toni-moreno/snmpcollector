package filter

import (
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/sirupsen/logrus"
	"github.com/soniah/gosnmp"
	"github.com/toni-moreno/snmpcollector/pkg/config"
)

// OidMultipleFilter a new Oid condition filter
type OidMultipleFilter struct {
	FilterLabels  map[string]string `json:"-"`
	EvalCondition string            //Eval condition
	condMap       map[string]*config.OidConditionCfg
	oidFilterMap  map[string]*OidFilter
	vars          []string
	//evaluable
	expr *govaluate.EvaluableExpression

	dbc  *config.DatabaseCfg
	Walk func(string, gosnmp.WalkFunc) error `json:"-"`
	log  *logrus.Logger
}

// NewOidMultipleFilter create a new filter for OID conditions
func NewOidMultipleFilter(cond string, l *logrus.Logger) *OidMultipleFilter {
	return &OidMultipleFilter{EvalCondition: cond, log: l}
}

// Init initialize
func (of *OidMultipleFilter) Init(arg ...interface{}) error {

	of.FilterLabels = make(map[string]string)
	of.condMap = make(map[string]*config.OidConditionCfg)

	of.Walk = arg[0].(func(string, gosnmp.WalkFunc) error)
	of.dbc = arg[1].(*config.DatabaseCfg)

	if of.Walk == nil {
		return fmt.Errorf("Error when initializing oid cond %s", of.EvalCondition)
	}
	if of.dbc == nil {
		return fmt.Errorf("Error when initializing oid cond %s", of.EvalCondition)
	}
	//needs to get data conditions
	expression, err := govaluate.NewEvaluableExpression(of.EvalCondition)
	if err != nil {
		of.log.Errorf("OIDMULTIPLEFILTER [%s] Error on initializing  ERROR : %s", of.EvalCondition, err)
		return err
	}
	of.expr = expression
	of.vars = of.expr.Vars()
	of.oidFilterMap = make(map[string]*OidFilter)

	for _, par := range of.vars {
		of.log.Infof("OIDMULTIPLEFILTER [%s] getting condition %s", of.EvalCondition, par)
		oidcond, err := of.dbc.GetOidConditionCfgByID(par)
		if err != nil {
			return err
		}
		of.condMap[par] = &oidcond
		of.log.Debugf("OIDMULTIPLEFILTER [%s]  get %s Filter %+v", of.EvalCondition, par, oidcond)
		f := NewOidFilter(oidcond.OIDCond, oidcond.CondType, oidcond.CondValue, of.log)
		f.Init(of.Walk)
		of.oidFilterMap[par] = f
	}

	return nil

}

// Count return current number of itemp in the filter
func (of *OidMultipleFilter) Count() int {
	return len(of.FilterLabels)
}

// MapLabels return the final tagmap from all posible values and the filter results
func (of *OidMultipleFilter) MapLabels(AllIndexedLabels map[string]string) map[string]string {
	curIndexedLabels := make(map[string]string, len(of.FilterLabels))
	for kf := range of.FilterLabels {
		for kl, vl := range AllIndexedLabels {
			if kf == kl {
				curIndexedLabels[kl] = vl
			}
		}
	}
	return curIndexedLabels
}

// Update use this to reload conditions
func (of *OidMultipleFilter) Update() error {
	filterMatrix := make(map[string]map[string]interface{}) //key [oidcond]
	for condID, f := range of.oidFilterMap {
		of.log.Debugf("OIDMULTIPLEFILTER [%s] updating filter data for key : %s", of.EvalCondition, condID)
		f.Update()
		for index := range f.FilterLabels {
			if _, ok := filterMatrix[index]; !ok {
				// not exist => create new var map and initialize
				a := make(map[string]interface{}, len(of.vars))
				for _, v := range of.vars {
					a[v] = false
				}
				a[condID] = true
				filterMatrix[index] = a
			} else {
				//already exist only set variable
				filterMatrix[index][condID] = true
			}
		}
	}
	//now we can already compute value
	of.FilterLabels = make(map[string]string)
	for kf, v := range filterMatrix {
		result, err := of.expr.Evaluate(v)
		if err != nil {
			of.log.Errorf("OIDMULTIPLEFILTER [%s] Error in expression evaluation for multiple filter On Index %s  ERROR : %s", of.EvalCondition, kf, err)
			continue
		}
		if result.(bool) {
			of.log.Debugf("OIDMULTIPLEFILTER [%s] Multiple filter expression for index %s TRUE on condition with map %+v", of.EvalCondition, kf, v)
			//save this index as true
			of.FilterLabels[kf] = ""
		}
	}
	return nil
}
