package filter

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/toni-moreno/snmpcollector/pkg/data/utils"
)

// FileFilter a filter loaded from a Config File placed in confDir (made for influxsnmp tool compatibility)
type FileFilter struct {
	filterLabels map[string]string
	FileName     string
	EnableAlias  bool
	confDir      string
	log          utils.Logger
}

// NewFileFilter creates a new Filter
func NewFileFilter(fileName string, enableAlias bool, l utils.Logger) *FileFilter {
	return &FileFilter{FileName: fileName, EnableAlias: enableAlias, log: l}
}

// Init load config first time
func (ff *FileFilter) Init(arg ...interface{}) error {
	ff.confDir = arg[0].(string)
	ff.log.Infof("FILEFILTER [%s] initialize File filter  Enable Alias: %t", ff.FileName, ff.EnableAlias)
	// chek if File exist
	if _, err := os.Stat(filepath.Join(ff.confDir, ff.FileName)); os.IsNotExist(err) {
		// does not exist
		ff.log.Errorf("FILEFILTER [%s] file %s does not exist Plase upload first to the %s dir: error: %s", ff.FileName, filepath.Join(ff.confDir, ff.FileName), ff.confDir, err)
		return err
	}
	ff.filterLabels = make(map[string]string)
	return nil
}

// Count return number of values in the filtered map
func (ff *FileFilter) Count() int {
	return len(ff.filterLabels)
}

// MapLabels return the final tagmap from all posible values and the filter results
func (ff *FileFilter) MapLabels(AllIndexedLabels map[string]string) map[string]string {
	curIndexedLabels := make(map[string]string, len(ff.filterLabels))
	for kf, vf := range ff.filterLabels {
		for kl, vl := range AllIndexedLabels {
			if kf == vl {
				if len(vf) > 0 {
					// map[k_l]v_f (alias to key of the label
					curIndexedLabels[kl] = vf
				} else {
					// map[k_l]v_l (original name)
					curIndexedLabels[kl] = vl
				}
			}
		}
	}
	return curIndexedLabels
}

// Update load filtered data from config file online time
func (ff *FileFilter) Update() error {
	ff.log.Infof("FILEFILTER [%s] apply File filter Enable Alias: %t", ff.FileName, ff.EnableAlias)
	// reset current fil ter
	ff.filterLabels = make(map[string]string)
	if len(ff.FileName) == 0 {
		return errors.New("No file configured error ")
	}
	data, err := ioutil.ReadFile(filepath.Join(ff.confDir, ff.FileName))
	if err != nil {
		ff.log.Errorf("FILEFILTER [%s] ERROR on open file %s: error: %s", ff.FileName, filepath.Join(ff.confDir, ff.FileName), err)
		return err
	}

	for lnum, line := range strings.Split(string(data), "\n") {
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
			ff.log.Warnf("FILEFILTER [%s] wrong number of parameters in file  Lnum: %d num : %d line: %s", ff.FileName, lnum, len(f), line)
		}
	}
	return nil
}
