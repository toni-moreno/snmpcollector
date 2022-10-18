package config

import (
	"fmt"
	"strconv"
)

func (dbc *DatabaseCfg) StartMigration() error {
	// snmpcollector - 0.14.0 - ensure new db table exists and its filled with current outputs
	ok, err := dbc.ValidateAndGenerateOutputTable()
	if err != nil {
		return err
	}
	if ok {
		log.Infof("MIGRATION: Migration succed, new outputs have been created")
		return nil
	}
	log.Infof("MIGRATION: Migration was skipped")
	return nil

}

func (dbc *DatabaseCfg) MigrateInfluxToOutputsDevs() error {
	session := dbc.x.NewSession()
	if err := session.Begin(); err != nil {
		// if returned then will rollback automatically
		return err
	}
	defer session.Close()

	// retrieve all devices:
	_, err := dbc.GetSnmpDeviceCfgArray("")
	if err != nil {
		session.Rollback()
		return err
	}

	return nil
}

func (dbc *DatabaseCfg) ValidateAndGenerateOutputTable() (bool, error) {
	log.Infof("MIGRATION: trying to migrate current InfluxDB servers to Outputs...")
	session := dbc.x.NewSession()
	if err := session.Begin(); err != nil {
		// if returned then will rollback automatically
		return false, err
	}
	defer session.Close()
	// check if new output_cfg table exists
	exists, err := session.IsTableExist("output_cfg")
	if err != nil {
		return false, err
	}
	if !exists {
		log.Infof("MIGRATION: failed to sync with output_cfg, skipping migration")
		return false, nil
	}
	// validate again to avoid some action on table
	n, err := dbc.GetOutputCfgArray("")
	if err != nil {
		return false, nil
	}
	if len(n) > 0 {
		log.Infof("MIGRATION: output_cfg exists and contains data, skipping migration")
		return false, nil
	}
	// retrieve the current influxdb configured servers
	icfgs, err := dbc.GetInfluxCfgArray("")
	if err != nil {
		return false, err
	}

	// retrieve the current influxdb configured servers with raw query
	// we need to retrieve the deprecated buffersize
	ricfgs, err := session.Query("select * from influx_cfg;")
	if err != nil {
		return false, err
	}
	var mod int
	for _, ricfg := range ricfgs {
		for iicfg, icfg := range icfgs {
			if icfg.ID == string(ricfg["id"]) {
				icfg.BufferSize, err = strconv.Atoi(string(ricfg["buffer_size"]))
				if err != nil {
					log.Errorf("Failed to retrieve buffer size, skipping output")
					//return false, err
				}
				icfgs[iicfg] = icfg
				mod++
			}
		}
	}

	if mod == 0 {
		return false, fmt.Errorf("no influxdb severs have been retrieved or legacy, skipping")
	}

	// influxdb outputs exists, so we need to update the current relation
	if len(n) == 0 && len(icfgs) > 0 {
		log.Infof("MIGRATION: Trying to create %d outputs based on influxdb servers", len(icfgs))
		for _, icfg := range icfgs {
			ocfg := &OutputCfg{
				ID:              icfg.ID,
				Active:          true,
				Backend:         icfg.ID,
				BufferSize:      icfg.BufferSize,
				FlushInterval:   60,
				MetricBatchSize: 6500,
				EnqueueOnError:  true,
			}
			_, err = session.Insert(ocfg)
			if err != nil {
				session.Rollback()
				return false, err
			}

			// add relations
			obktruct := OutputBackends{
				IDOutput:    icfg.ID,
				IDBackend:   icfg.ID,
				BackendType: "influxdb",
			}
			_, err := session.Insert(&obktruct)
			if err != nil {
				session.Rollback()
				return false, err
			}
		}
	}
	err = session.Commit()
	if err != nil {
		return false, err
	}
	return true, nil
}
