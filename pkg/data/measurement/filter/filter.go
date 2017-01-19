package filter

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/soniah/gosnmp"
	"github.com/toni-moreno/snmpcollector/pkg/data/snmp"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	confDir string //Needed to get File Filters measurments
)

// SetConfDir  enable load File Filters from anywhere in the our FS.
func SetConfDir(dir string) {
	confDir = dir
}

type Filter interface {
	LoadCfg() error
	Compute(arg ...interface{}) error
	//construct the final index array from all index and filters
	MapLabels(map[string]string) map[string]string
	Count() int
}

type FileFilter struct {
	filterLabels map[string]string
	FileName     string
	EnableAlias  bool
	log          *logrus.Logger
}

func NewFileFilter(fileName string, enableAlias bool, l *logrus.Logger) *FileFilter {
	return &FileFilter{FileName: fileName, EnableAlias: enableAlias, log: l}
}

func (ff *FileFilter) LoadCfg() error {
	return nil
}

func (ff *FileFilter) Count() int {
	return len(ff.filterLabels)
}

func (ff *FileFilter) MapLabels(AllIndexedLabels map[string]string) map[string]string {
	curIndexedLabels := make(map[string]string, len(ff.filterLabels))
	for k_f, v_f := range ff.filterLabels {
		for k_l, v_l := range AllIndexedLabels {
			if k_f == v_l {
				if len(v_f) > 0 {
					// map[k_l]v_f (alias to key of the label
					curIndexedLabels[k_l] = v_f
				} else {
					//map[k_l]v_l (original name)
					curIndexedLabels[k_l] = v_l
				}

			}
		}
	}
	return curIndexedLabels
}

func (ff *FileFilter) Compute(arg ...interface{}) error {
	ff.log.Infof("apply File filter : %s Enable Alias: %t", ff.FileName, ff.EnableAlias)
	ff.filterLabels = make(map[string]string)
	if len(ff.FileName) == 0 {
		return errors.New("No file configured error ")
	}
	data, err := ioutil.ReadFile(filepath.Join(confDir, ff.FileName))
	if err != nil {
		ff.log.Errorf("ERROR on open file %s: error: %s", filepath.Join(confDir, ff.FileName), err)
		return err
	}

	for l_num, line := range strings.Split(string(data), "\n") {
		//		log.Println("LINIA:", line)
		// strip comments
		comment := strings.Index(line, "#")
		if comment >= 0 {
			line = line[:comment]
		}
		if len(line) == 0 {
			continue
		}
		f := strings.Fields(line)
		switch len(f) {
		case 1:
			ff.filterLabels[f[0]] = ""

		case 2:
			if ff.EnableAlias {
				ff.filterLabels[f[0]] = f[1]
			} else {
				ff.filterLabels[f[0]] = ""
			}

		default:
			ff.log.Warnf("wrong number of parameters in file: %s Lnum: %s num : %s line: %s", ff.FileName, l_num, len(f), line)
		}
	}
	return nil
}

// OidFilter a new Oid condition filter
type OidFilter struct {
	filterLabels map[string]string
	OidCond      string
	TypeCond     string
	ValueCond    string
	log          *logrus.Logger
}

func NewOidFilter(oidcond string, typecond string, value string, l *logrus.Logger) *OidFilter {
	return &OidFilter{OidCond: oidcond, TypeCond: typecond, ValueCond: value, log: l}
}

func (of *OidFilter) LoadCfg() error {
	return nil
}

func (of *OidFilter) Count() int {
	return len(of.filterLabels)
}

func (of *OidFilter) MapLabels(AllIndexedLabels map[string]string) map[string]string {
	curIndexedLabels := make(map[string]string, len(of.filterLabels))
	for k_f, _ := range of.filterLabels {
		for k_l, v_l := range AllIndexedLabels {
			if k_f == k_l {
				curIndexedLabels[k_l] = v_l
			}
		}
	}
	return curIndexedLabels
}

func (of *OidFilter) Compute(arg ...interface{}) error {
	Walk := arg[0].(func(string, gosnmp.WalkFunc) error)
	of.log.Debugf("Compute Condition Filter: Looking up column names in: Condition %s", of.OidCond)

	idxPosInOID := len(of.OidCond)

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
	err := Walk(of.OidCond, setRawData)
	if err != nil {
		of.log.Errorf("SNMP  walk error : %s", err)
		return err
	}

	return nil
}
