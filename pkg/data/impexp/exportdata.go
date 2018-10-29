package impexp

import (
	"fmt"

	"time"

	"github.com/Knetic/govaluate"
	"github.com/Sirupsen/logrus"
	"snmpcollector/pkg/agent"
	"snmpcollector/pkg/config"
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

// ExportInfo Main export Data type
type ExportInfo struct {
	FileName      string
	Description   string
	Author        string
	Tags          string
	AgentVersion  string
	ExportVersion string
	CreationDate  time.Time
}

// EIOptions export/import options
type EIOptions struct {
	Recursive   bool   //Export Option
	AutoRename  bool   //Import Option
	AlternateID string //Import Option
}

// ExportObject Base type for any object to export
type ExportObject struct {
	ObjectTypeID string
	ObjectID     string
	Options      *EIOptions
	ObjectCfg    interface{}
	Error        string
}

// ExportData the runtime measurement config
type ExportData struct {
	Info       *ExportInfo
	Objects    []*ExportObject
	tmpObjects []*ExportObject //only for temporal use
}

// NewExport ExportData type creator
func NewExport(info *ExportInfo) *ExportData {
	if len(agent.Version) > 0 {
		info.AgentVersion = agent.Version
	} else {
		info.AgentVersion = "debug"
	}

	info.ExportVersion = "1.0"
	info.CreationDate = time.Now()
	return &ExportData{
		Info: info,
	}
}

func checkIfExistOnArray(list []*ExportObject, ObjType string, id string) bool {
	for _, v := range list {
		if v.ObjectTypeID == ObjType && v.ObjectID == id {
			return true
		}
	}
	return false
}

// PrependObject prepend a new object to the ExportData type
func (e *ExportData) PrependObject(obj *ExportObject) {
	if checkIfExistOnArray(e.Objects, obj.ObjectTypeID, obj.ObjectID) == true {
		return
	}
	e.tmpObjects = append([]*ExportObject{obj}, e.tmpObjects...)
}

// UpdateTmpObject update temporaty object
func (e *ExportData) UpdateTmpObject() {
	//we need remove duplicated objects on the auxiliar array
	objectList := []*ExportObject{}
	for i := 0; i < len(e.tmpObjects); i++ {
		v := e.tmpObjects[i]
		if checkIfExistOnArray(objectList, v.ObjectTypeID, v.ObjectID) == false {
			objectList = append(objectList, v)
		}
	}
	e.Objects = append(e.Objects, objectList...)
	e.tmpObjects = nil
}

// Export  exports data
func (e *ExportData) Export(ObjType string, id string, recursive bool, level int) error {

	switch ObjType {
	case "snmpdevicecfg":
		//contains sensible data
		v, err := dbc.GetSnmpDeviceCfgByID(id)
		if err != nil {
			return err
		}
		e.PrependObject(&ExportObject{ObjectTypeID: "snmpdevicecfg", ObjectID: id, ObjectCfg: v})
		if !recursive {
			break
		}
		for _, val := range v.MeasurementGroups {
			e.Export("measgroupcfg", val, recursive, level+1)
		}
		for _, val := range v.MeasFilters {
			e.Export("measfiltercfg", val, recursive, level+1)
		}
		e.Export("influxcfg", v.OutDB, recursive, level+1)
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
		if !recursive {
			break
		}
		switch v.FType {
		case "file":
		case "OIDCondition":
			e.Export("oidconditioncfg", v.FilterName, recursive, level+1)
		case "CustomFilter":
			e.Export("customfiltercfg", v.FilterName, recursive, level+1)
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
		if !recursive {
			break
		}
		if v.IsMultiple {
			expression, err := govaluate.NewEvaluableExpression(v.OIDCond)
			if err != nil {
				return fmt.Errorf("Error on initializing , evaluation : %s : ERROR : %s", v.OIDCond, err)
			}
			vars := expression.Vars()
			for _, par := range vars {
				oidcond, err := dbc.GetOidConditionCfgByID(par)
				if err != nil {
					return fmt.Errorf("Error on initializing , evaluation : %s (subcondition %s): ERROR : %s", v.OIDCond, par, err)
				}
				//TODO review if this should be a recursive export better than prepend
				e.PrependObject(&ExportObject{ObjectTypeID: "oidconditioncfg", ObjectID: par, ObjectCfg: oidcond})
			}
		}
	case "measurementcfg":
		v, err := dbc.GetMeasurementCfgByID(id)
		if err != nil {
			return err
		}
		e.PrependObject(&ExportObject{ObjectTypeID: "measurementcfg", ObjectID: id, ObjectCfg: v})
		if !recursive {
			break
		}
		//--------------------
		// metric objects
		//--------------------

		for _, val := range v.Fields {
			e.Export("snmpmetriccfg", val.ID, recursive, level+1)
		}

		//--------------------
		// Var Catalog Objects
		//--------------------
		// we need to initialice the measurment first to get Variables.

		cfg, _ := dbc.GetSnmpMetricCfgMap("")
		gv, _ := dbc.GetVarCatalogCfgMap("")

		err = v.Init(&cfg, config.CatalogVar2Map(gv))
		if err != nil {
			return err
		}
		//now we can get Used Vars
		vara, err := v.GetExternalVars()
		if err != nil {
			log.Warnf("There is some problem while trying to get variables used in this measurement: %s", err)
		}
		log.Debugf("GET EXTERNAL VARS in  measurment %s: VARS :%+v", v.ID, vara)
		// var array to object array
		var varca []config.VarCatalogCfg
		for _, val := range vara {
			log.Debugf("Workign with variables: %s", val)
			v, err := dbc.GetVarCatalogCfgByID(val)
			if err != nil {
				return err
			}
			varca = append(varca, v)
		}
		// var catalong objects
		for _, val := range varca {
			e.Export("varcatalogcfg", val.ID, recursive, level+1)
		}
	case "snmpmetriccfg":
		v, err := dbc.GetSnmpMetricCfgByID(id)
		if err != nil {
			return err
		}
		e.PrependObject(&ExportObject{ObjectTypeID: "snmpmetriccfg", ObjectID: id, ObjectCfg: v})
		if v.DataSrcType == "CONDITIONEVAL" {
			e.Export("oidconditioncfg", v.ExtraData, recursive, level+1)
		}
	case "measgroupcfg":
		v, err := dbc.GetMGroupsCfgByID(id)
		if err != nil {
			return err
		}
		e.PrependObject(&ExportObject{ObjectTypeID: "measgroupcfg", ObjectID: id, ObjectCfg: v})
		if !recursive {
			break
		}
		for _, val := range v.Measurements {
			e.Export("measurementcfg", val, recursive, level+1)
		}
	case "varcatalogcfg":
		v, err := dbc.GetVarCatalogCfgByID(id)
		if err != nil {
			return err
		}
		e.PrependObject(&ExportObject{ObjectTypeID: "varcatalogcfg", ObjectID: id, ObjectCfg: v})
	default:
		return fmt.Errorf("Unknown type object type %s ", ObjType)
	}
	if level == 0 {
		e.UpdateTmpObject()
	}
	return nil
}
