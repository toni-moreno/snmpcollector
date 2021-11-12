package filter

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gosnmp/gosnmp"
	"github.com/toni-moreno/snmpcollector/pkg/data/snmp"
	"github.com/toni-moreno/snmpcollector/pkg/data/utils"
)

// OidFilter a new Oid condition filter
type OidFilter struct {
	FilterLabels map[string]string `json:"-"`
	OidCond      string
	TypeCond     string
	ValueCond    string
	log          utils.Logger
	Walk         func(string, gosnmp.WalkFunc) error `json:"-"`
}

// NewOidFilter create a new filter for OID conditions
func NewOidFilter(oidcond string, typecond string, value string, l utils.Logger) *OidFilter {
	return &OidFilter{OidCond: oidcond, TypeCond: typecond, ValueCond: value, log: l}
}

// Init initialize
func (of *OidFilter) Init(arg ...interface{}) error {
	of.FilterLabels = make(map[string]string)
	of.Walk = arg[0].(func(string, gosnmp.WalkFunc) error)
	if of.Walk == nil {
		return fmt.Errorf("Error when initializing oid cond %s", of.OidCond)
	}
	return nil
}

// Count return current number of itemp in the filter
func (of *OidFilter) Count() int {
	return len(of.FilterLabels)
}

// MapLabels return the final tagmap from all posible values and the filter results
func (of *OidFilter) MapLabels(AllIndexedLabels map[string]string) map[string]string {
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

// Update  load filtered data from SNMP client config online time
func (of *OidFilter) Update() error {
	of.log.Debugf("OIDFILTER [%s] Compute Condition Filter: Looking up column names in Condition", of.OidCond)

	idxPosInOID := len(of.OidCond)
	// reset current filter
	of.FilterLabels = make(map[string]string)

	setRawData := func(pdu gosnmp.SnmpPDU) error {
		of.log.Debugf("OIDFILTER [%s] received SNMP  pdu:%+v", of.OidCond, pdu)
		if pdu.Value == nil {
			of.log.Warnf("OIDFILTER [%s] no value retured by pdu :%+v", of.OidCond, pdu)
			return nil // if error return the bulk process will stop
		}
		var vci int64
		var value int64
		var cond bool

		switch {
		case of.TypeCond == "notmatch":
			// m.log.Debugf("PDU: %+v", pdu)
			str := snmp.PduVal2str(pdu)
			re, err := regexp.Compile(of.ValueCond)
			if err != nil {
				of.log.Warnf("OIDFILTER [%s] Evaluated notmatch condition  value: %s | filter: %s | ERROR : %s", of.OidCond, str, of.ValueCond, err)
				break
			}
			matched := re.MatchString(str)
			of.log.Debugf("OIDFILTER [%s] Evaluated notmatch condition  value: %s | filter: %s | result : %t", of.OidCond, str, of.ValueCond, !matched)
			cond = !matched
		case of.TypeCond == "match":
			// m.log.Debugf("PDU: %+v", pdu)
			str := snmp.PduVal2str(pdu)
			re, err := regexp.Compile(of.ValueCond)
			if err != nil {
				of.log.Warnf("OIDFILTER [%s] Evaluated match condition  value: %s | filter: %s | ERROR : %s", of.OidCond, str, of.ValueCond, err)
				break
			}
			matched := re.MatchString(str)
			of.log.Debugf("OIDFILTER [%s] Evaluated match condition  value: %s | filter: %s | result : %t", of.OidCond, str, of.ValueCond, matched)
			cond = matched
		case of.TypeCond == "nin":
			// Numeric In
			iarray, err := utils.CSV2IntArray(of.ValueCond)
			if err != nil {
				of.log.Warnf("OIDFILTER [%s] error on CSV to IntegerArray accepted numeric value as value condition  current : %s  for TypeCond %s: Error: %s", of.OidCond, of.ValueCond, of.TypeCond, err)
				return nil
			}
			value = snmp.PduVal2Int64(pdu)
			cond = false
			for _, v := range iarray {
				cond = (value == v)
				if cond {
					break
				}
			}
		case strings.Contains(of.TypeCond, "n"):
			// undesrstand valueCondition as numeric
			vc, err := strconv.Atoi(of.ValueCond)
			if err != nil {
				of.log.Warnf("OIDFILTER [%s] only accepted numeric value as value condition  current : %s  for TypeCond %s", of.OidCond, of.ValueCond, of.TypeCond)
				return nil
			}
			vci = int64(vc)
			// TODO review types
			value = snmp.PduVal2Int64(pdu)
			switch {
			case of.TypeCond == "neq":
				cond = (value == vci)
				of.log.Debugf("OIDFILTER [%s] type [%s] snmp value [%d], condition [%d] RESULT[%t]", of.OidCond, of.TypeCond, value, vci, cond)
			case of.TypeCond == "nlt":
				cond = (value < vci)
				of.log.Debugf("OIDFILTER [%s] type [%s] snmp value [%d], condition [%d] RESULT[%t]", of.OidCond, of.TypeCond, value, vci, cond)
			case of.TypeCond == "ngt":
				cond = (value > vci)
				of.log.Debugf("OIDFILTER [%s] type [%s] snmp value [%d], condition [%d] RESULT[%t]", of.OidCond, of.TypeCond, value, vci, cond)
			case of.TypeCond == "nge":
				cond = (value >= vci)
				of.log.Debugf("OIDFILTER [%s] type [%s] snmp value [%d], condition [%d] RESULT[%t]", of.OidCond, of.TypeCond, value, vci, cond)
			case of.TypeCond == "nle":
				cond = (value <= vci)
				of.log.Debugf("OIDFILTER [%s] type [%s] snmp value [%d], condition [%d] RESULT[%t]", of.OidCond, of.TypeCond, value, vci, cond)
			case of.TypeCond == "ndif":
				cond = (value != vci)
				of.log.Debugf("OIDFILTER [%s] type [%s] snmp value [%d], condition [%d] RESULT[%t]", of.OidCond, of.TypeCond, value, vci, cond)

			}
		default:
			of.log.Errorf("OIDFILTER [%s] Error in Condition filter  Type: %s ValCond: %s ", of.OidCond, of.TypeCond, of.ValueCond)
		}
		if cond == true {
			if len(pdu.Name) < idxPosInOID+1 {
				of.log.Warnf("OIDFILTER [%s] Received PDU OID smaller  than minimal index(%d) positionretured by pdu :%+v", of.OidCond, idxPosInOID, pdu)
				return nil // if error return the bulk process will stop
			}
			suffix := pdu.Name[idxPosInOID+1:]
			of.FilterLabels[suffix] = ""
		}

		return nil
	}
	err := of.Walk(of.OidCond, setRawData)
	if err != nil {
		of.log.Errorf("OIDFILTER [%s] SNMP  walk error : %s", of.OidCond, err)
		return err
	}

	return nil
}
