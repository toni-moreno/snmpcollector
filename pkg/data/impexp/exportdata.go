package impexp

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/toni-moreno/snmpcollector/pkg/agent"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"time"
)

var (
	log     *logrus.Logger
	confDir string              //Needed to get File Filters data
	dbc     *config.DatabaseCfg //Needed to get Custom Filter  data
)

// SetConfDir  enable load File Filters from anywhere in the our FS.
func SetConfDir(dir string) {
	confDir = dir
}

// SetDB load database config to load data if needed (used in filters)
func SetDB(db *config.DatabaseCfg) {
	dbc = db
}

// SetLogger set output log
func SetLogger(l *logrus.Logger) {
	log = l
}

type ExportInfo struct {
	FileName    string
	Description string
	Author      string
	Tags        string
}

type ExportObject struct {
	ObjectTypeID string
	ObjectID     string
	ObjectCfg    interface{}
}

// ExportData the runtime measurement config
type ExportData struct {
	Info          *ExportInfo
	AgentVersion  string
	ExportVersion string
	CreationDate  time.Time
	Objects       []*ExportObject
}

func NewExport(info *ExportInfo) *ExportData {
	return &ExportData{
		Info:          info,
		AgentVersion:  agent.Version,
		ExportVersion: "1.0",
		CreationDate:  time.Now(),
	}
}

func (e *ExportData) checkIfExist(ObjType string, id string) bool {
	for _, v := range e.Objects {
		if v.ObjectTypeID == ObjType && v.ObjectID == id {
			return true
		}
	}
	return false
}

func (e *ExportData) PrependObject(obj *ExportObject) {
	if e.checkIfExist(obj.ObjectTypeID, obj.ObjectID) {
		return
	}
	e.Objects = append([]*ExportObject{obj}, e.Objects...)
}

// Export  exports data
func (e *ExportData) Export(ObjType string, id string) error {

	switch ObjType {
	case "snmpdevicecfg":
		//contains sensible data
		v, err := dbc.GetSnmpDeviceCfgByID(id)
		if err != nil {
			return err
		}
		e.PrependObject(&ExportObject{ObjectTypeID: "snmpdevicecfg", ObjectID: id, ObjectCfg: v})
		for _, val := range v.MeasurementGroups {
			e.Export("measgroupcfg", val)
		}
		for _, val := range v.MeasFilters {
			e.Export("measfiltercfg", val)
		}
		e.Export("influxcfg", v.OutDB)
	case "influxcfg":
		//contains sensible probable
		v, err := dbc.GetInfluxCfgByID(id)
		if err != nil {
			return err
		}
		e.PrependObject(&ExportObject{ObjectTypeID: "influxcfg", ObjectID: id, ObjectCfg: v})
	case "measfiltercfg":
		v, err := dbc.GetMeasFilterCfgByID(id)
		if err != nil {
			return err
		}
		e.PrependObject(&ExportObject{ObjectTypeID: "measfiltercfg", ObjectID: id, ObjectCfg: v})
		switch v.FType {
		case "file":
		case "OIDCondition":
			e.Export("oidconditioncfg", v.FilterName)
		case "CustomFilter":
			e.Export("customfiltercfg", v.FilterName)
		}
	case "customfiltercfg":
		v, err := dbc.GetCustomFilterCfgByID(id)
		if err != nil {
			return err
		}
		e.PrependObject(&ExportObject{ObjectTypeID: "customfiltercfg", ObjectID: id, ObjectCfg: v})
	case "oidconditioncfg":
		v, err := dbc.GetOidConditionCfgByID(id)
		if err != nil {
			return err
		}
		e.PrependObject(&ExportObject{ObjectTypeID: "oidconditioncfg", ObjectID: id, ObjectCfg: v})
	case "measurementcfg":
		v, err := dbc.GetMeasurementCfgByID(id)
		if err != nil {
			return err
		}
		e.PrependObject(&ExportObject{ObjectTypeID: "measurementcfg", ObjectID: id, ObjectCfg: v})
		for _, val := range v.Fields {
			e.Export("snmpmetriccfg", val.ID)
		}
	case "snmpmetriccfg":
		v, err := dbc.GetSnmpMetricCfgByID(id)
		if err != nil {
			return err
		}
		e.PrependObject(&ExportObject{ObjectTypeID: "snmpmetriccfg", ObjectID: id, ObjectCfg: v})
		if v.DataSrcType == "CONDITIONEVAL" {
			e.Export("oidconditioncfg", v.ExtraData)
		}
	case "measgroupcfg":
		v, err := dbc.GetMGroupsCfgByID(id)
		if err != nil {
			return err
		}
		e.PrependObject(&ExportObject{ObjectTypeID: "measgroupscfg", ObjectID: id, ObjectCfg: v})
		for _, val := range v.Measurements {
			e.Export("measurementcfg", val)
		}

	default:
		return fmt.Errorf("Unknown type object type %s ", ObjType)
	}

	return nil
}
