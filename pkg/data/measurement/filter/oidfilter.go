package filter

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/soniah/gosnmp"
	"github.com/toni-moreno/snmpcollector/pkg/data/snmp"
	"regexp"
	"strconv"
	"strings"
)

// OidFilter a new Oid condition filter
type OidFilter struct {
	filterLabels map[string]string
	OidCond      string
	TypeCond     string
	ValueCond    string
	log          *logrus.Logger
	Walk         func(string, gosnmp.WalkFunc) error `json:"-"`
}

// NewOidFilter create a new filter for OID conditions
func NewOidFilter(oidcond string, typecond string, value string, l *logrus.Logger) *OidFilter {
	return &OidFilter{OidCond: oidcond, TypeCond: typecond, ValueCond: value, log: l}
}

// Init initialize
func (of *OidFilter) Init(arg ...interface{}) error {
	of.filterLabels = make(map[string]string)
	of.Walk = arg[0].(func(string, gosnmp.WalkFunc) error)
	if of.Walk == nil {
		return fmt.Errorf("Error when initializing oid cond %s", of.OidCond)
	}
	return nil
}

// Count return current number of itemp in the filter
func (of *OidFilter) Count() int {
	return len(of.filterLabels)
}

// MapLabels return the final tagmap from all posible values and the filter results
func (of *OidFilter) MapLabels(AllIndexedLabels map[string]string) map[string]string {
	curIndexedLabels := make(map[string]string, len(of.filterLabels))
	for kf := range of.filterLabels {
		for kl, vl := range AllIndexedLabels {
			if kf == kl {
				curIndexedLabels[kl] = vl
			}
		}
	}
	return curIndexedLabels
}

// Update  load filtered data from SNMP client config online time
func (of *OidFilter) Update() error {

	of.log.Debugf("Compute Condition Filter: Looking up column names in: Condition %s", of.OidCond)

	idxPosInOID := len(of.OidCond)
	// reset current filter
	of.filterLabels = make(map[string]string)

	setRawData := func(pdu gosnmp.SnmpPDU) error {
		of.log.Debugf("received SNMP  pdu:%+v", pdu)
		if pdu.Value == nil {
			of.log.Warnf("no value retured by pdu :%+v", pdu)
			return nil //if error return the bulk process will stop
		}
		var vci int64
		var value int64
		var cond bool

		switch {
		case of.TypeCond == "notmatch":
			//m.log.Debugf("PDU: %+v", pdu)
			str := snmp.PduVal2str(pdu)
			var re = regexp.MustCompile(of.ValueCond)
			matched := re.MatchString(str)
			of.log.Debugf("Evaluated notmatch condition  value: %s | filter: %s | result : %t", str, of.ValueCond, !matched)
			cond = !matched
		case of.TypeCond == "match":
			//m.log.Debugf("PDU: %+v", pdu)
			str := snmp.PduVal2str(pdu)
			var re = regexp.MustCompile(of.ValueCond)
			matched := re.MatchString(str)
			of.log.Debugf("Evaluated match condition  value: %s | filter: %s | result : %t", str, of.ValueCond, matched)
			cond = matched
		case strings.Contains(of.TypeCond, "n"):
			//undesrstand valueCondition as numeric
			vc, err := strconv.Atoi(of.ValueCond)
			if err != nil {
				of.log.Warnf("only accepted numeric value as value condition  current : %s  for TypeCond %s", of.ValueCond, of.TypeCond)
				return nil
			}
			vci = int64(vc)
			//TODO review types
			value = snmp.PduVal2Int64(pdu)
			fallthrough
		case of.TypeCond == "neq":
			cond = (value == vci)
		case of.TypeCond == "nlt":
			cond = (value < vci)
		case of.TypeCond == "ngt":
			cond = (value > vci)
		case of.TypeCond == "nge":
			cond = (value >= vci)
		case of.TypeCond == "nle":
			cond = (value <= vci)
		default:
			of.log.Errorf("Error in Condition filter OidCondition: %s Type: %s ValCond: %s ", of.OidCond, of.TypeCond, of.ValueCond)
		}
		if cond == true {
			if len(pdu.Name) < idxPosInOID {
				of.log.Warnf("Received PDU OID smaller  than minimal index(%d) positionretured by pdu :%+v", idxPosInOID, pdu)
				return nil //if error return the bulk process will stop
			}
			suffix := pdu.Name[idxPosInOID+1:]
			of.filterLabels[suffix] = ""
		}

		return nil
	}
	err := of.Walk(of.OidCond, setRawData)
	if err != nil {
		of.log.Errorf("SNMP  walk error : %s", err)
		return err
	}

	return nil
}
