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
	ObjectPtr    interface{}
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
	//newdata := []*ExportObject{obj}
	//e.Objects = append(newdata, e.Objects...)
	e.Objects = append([]*ExportObject{obj}, e.Objects...)
}

func (e *ExportData) ExportPtr(ptr interface{}) error {

	switch v := ptr.(type) {
	case config.SnmpDeviceCfg:
		//contains sensible data
	case config.InfluxCfg:
		//contains sensible probable
	case config.MeasFilterCfg:
		e.PrependObject(&ExportObject{ObjectTypeID: "measfiltercfg", ObjectID: v.ID, ObjectPtr: ptr})
		switch v.FType {
		case "file":
		case "OIDCondition":
			filter, _ := dbc.GetOidConditionCfgByID(v.FilterName)
			e.ExportPtr(filter)
			//e.PrependObject(&ExportObject{ObjectTypeID: "oidconditioncfg", ObjectID: v.FilterName, ObjectPtr: filter})
		case "CustomFilter":
			filter, _ := dbc.GetCustomFilterCfgByID(v.FilterName)
			e.ExportPtr(filter)
			//e.PrependObject(&ExportObject{ObjectTypeID: "customfiltercfg", ObjectID: v.FilterName, ObjectPtr: filter})
		}

	case config.CustomFilterCfg:
		e.PrependObject(&ExportObject{ObjectTypeID: "customfiltercfg", ObjectID: v.ID, ObjectPtr: ptr})

	case config.OidConditionCfg:
		e.PrependObject(&ExportObject{ObjectTypeID: "oidconditioncfg", ObjectID: v.ID, ObjectPtr: ptr})

	case config.MeasurementCfg:
		e.PrependObject(&ExportObject{ObjectTypeID: "measurementcfg", ObjectID: v.ID, ObjectPtr: ptr})
		for _, val := range v.Fields {
			metric, _ := dbc.GetSnmpMetricCfgByID(val.ID)
			e.ExportPtr(metric)
		}
	case config.SnmpMetricCfg:
		e.PrependObject(&ExportObject{ObjectTypeID: "snmpmetriccfg", ObjectID: v.ID, ObjectPtr: ptr})
		if v.DataSrcType == "CONDITIONEVAL" {
			cond, _ := dbc.GetOidConditionCfgByID(v.ExtraData)
			e.ExportPtr(cond)
		}
	case config.MGroupsCfg:
		e.PrependObject(&ExportObject{ObjectTypeID: "measgroupscfg", ObjectID: v.ID, ObjectPtr: ptr})
		for _, val := range v.Measurements {
			meas, _ := dbc.GetMeasurementCfgByID(val)
			e.ExportPtr(meas)
		}

	default:
		log.Warn("Unknown type for object ")
	}

	return nil
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
		e.PrependObject(&ExportObject{ObjectTypeID: "snmpdevicecfg", ObjectID: id, ObjectPtr: v})
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
		e.PrependObject(&ExportObject{ObjectTypeID: "influxcfg", ObjectID: id, ObjectPtr: v})

	case "measfiltercfg":
		v, err := dbc.GetMeasFilterCfgByID(id)
		if err != nil {
			return err
		}
		e.PrependObject(&ExportObject{ObjectTypeID: "measfiltercfg", ObjectID: id, ObjectPtr: v})
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
		e.PrependObject(&ExportObject{ObjectTypeID: "customfiltercfg", ObjectID: id, ObjectPtr: v})

	case "oidconditioncfg":
		v, err := dbc.GetOidConditionCfgByID(id)
		if err != nil {
			return err
		}
		e.PrependObject(&ExportObject{ObjectTypeID: "oidconditioncfg", ObjectID: id, ObjectPtr: v})

	case "measurementcfg":
		v, err := dbc.GetMeasurementCfgByID(id)
		if err != nil {
			return err
		}
		e.PrependObject(&ExportObject{ObjectTypeID: "measurementcfg", ObjectID: id, ObjectPtr: v})
		for _, val := range v.Fields {
			e.Export("snmpmetriccfg", val.ID)
		}
	case "snmpmetriccfg":
		v, err := dbc.GetSnmpMetricCfgByID(id)
		if err != nil {
			return err
		}
		e.PrependObject(&ExportObject{ObjectTypeID: "snmpmetriccfg", ObjectID: id, ObjectPtr: v})
		if v.DataSrcType == "CONDITIONEVAL" {
			e.Export("oidconditioncfg", v.ExtraData)
		}
	case "measgroupcfg":
		v, err := dbc.GetMGroupsCfgByID(id)
		if err != nil {
			return err
		}
		e.PrependObject(&ExportObject{ObjectTypeID: "measgroupscfg", ObjectID: id, ObjectPtr: v})
		for _, val := range v.Measurements {
			e.Export("measurementcfg", val)
		}

	default:
		return fmt.Errorf("Unknown type object type %s ", ObjType)
	}

	return nil
}
