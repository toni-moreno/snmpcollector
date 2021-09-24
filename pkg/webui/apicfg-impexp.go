package webui

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"mime/multipart"
	"os"

	"github.com/go-macaron/binding"
	"github.com/toni-moreno/snmpcollector/pkg/data/impexp"
	"gopkg.in/macaron.v1"
)

// UploadForm form struct
type UploadForm struct {
	AutoRename bool
	OverWrite  bool
	ExportFile *multipart.FileHeader
}

// NewAPICfgImportExport Import/Export REST API creator
func NewAPICfgImportExport(m *macaron.Macaron) error {
	bind := binding.Bind

	m.Group("/api/cfg/export", func() {
		m.Get("/:objtype/:id", reqSignedIn, ExportObject)
		m.Post("/:objtype/:id", reqSignedIn, bind(impexp.ExportInfo{}), ExportObjectToFile)
	})

	m.Group("/api/cfg/bulkexport", func() {
		m.Post("/", reqSignedIn, binding.Json(impexp.ExportData{}), BulkExportObjectToFile)
	})

	m.Group("/api/cfg/import", func() {
		m.Post("/", reqSignedIn, binding.MultipartForm(UploadForm{}), ImportDataFile)
	})
	return nil
}

/****************/
/*IMPORT*/
/*****************/

// ImportCheck import check struct
type ImportCheck struct {
	IsOk    bool
	Message string
	Data    *impexp.ExportData
}

// ImportDataFile import data from uploaded file
func ImportDataFile(ctx *Context, uf UploadForm) {
	if (UploadForm{}) == uf {
		log.Error("Error no data in expected struct")
		ctx.JSON(404, "Error no data in expected struct")
		return
	}
	log.Debugf("Uploaded data :%+v", uf)
	if uf.ExportFile == nil {
		ctx.JSON(404, "Error no file uploaded struct")
		return
	}
	log.Debugf("Uploaded File : %+v", uf)
	file, err := uf.ExportFile.Open()
	if err != nil {
		log.Warningf("Error on Open Uploaded File: %s", err)
		ctx.JSON(404, err.Error())
		return
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(file)
	s := buf.String()
	log.Debugf("FILE DATA: %s", s)
	ImportedData := impexp.ExportData{}
	if err := json.Unmarshal(buf.Bytes(), &ImportedData); err != nil {
		log.Errorf("Error in data to struct (json-unmarshal) procces: %s", err)
		ctx.JSON(404, err.Error())
		return
	}
	log.Debugf("IMPORTED STRUCT %+v", ImportedData)

	a, err := ImportedData.ImportCheck()

	if err != nil && uf.AutoRename == false && uf.OverWrite == false {
		ctx.JSON(200, &ImportCheck{IsOk: false, Message: err.Error(), Data: a})
		return
	}
	err = ImportedData.Import(uf.OverWrite, uf.AutoRename)
	if err != nil {
		log.Errorf("Some Error happened on import data: %s", err)
		ctx.JSON(200, &ImportCheck{IsOk: false, Message: err.Error(), Data: &ImportedData})
		return
	}
	ctx.JSON(200, &ImportCheck{IsOk: true, Message: "all objects have been  imported", Data: &ImportedData})
}

/****************/
/*EXPORT*/
/****************/

// ExportObject export object
func ExportObject(ctx *Context) {
	id := ctx.Params(":id")
	objtype := ctx.Params(":objtype")
	exp := impexp.NewExport(&impexp.ExportInfo{FileName: "autogenerated.txt", Description: "autogenerated"})
	err := exp.Export(objtype, id, true, 0)
	if err != nil {
		log.Warningf("Error on get object array for type %s with ID %s  , error: %s", objtype, id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &exp)
	}
}

func generateFile(ctx *Context, exp *impexp.ExportData) {
	outdata, err := json.Marshal(exp)
	if err != nil {
		log.Errorf("Error on Json data formatting: %s", err)
		ctx.JSON(404, err.Error())
		return
	}
	tmpfile, err := ioutil.TempFile("", "snmpcollector")
	if err != nil {
		log.Errorf("Error on create new temporal file: %s", err)
		ctx.JSON(404, err.Error())
		return
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write(outdata); err != nil {
		log.Errorf("Error on write data to temporal file: %s", tmpfile.Name())
		ctx.JSON(404, err.Error())
		return
	}
	if err := tmpfile.Close(); err != nil {
		log.Errorf("Error on close temporal file: %s", tmpfile.Name())
		ctx.JSON(404, err.Error())
		return
	}

	ctx.ServeFile(tmpfile.Name())
}

// ExportObjectToFile export Object to file
func ExportObjectToFile(ctx *Context, info impexp.ExportInfo) {
	id := ctx.Params(":id")
	objtype := ctx.Params(":objtype")
	exp := impexp.NewExport(&info)
	err := exp.Export(objtype, id, true, 0)
	if err != nil {
		log.Errorf("Error on create Export Data with: %s", err)
		ctx.JSON(404, err.Error())
		return
	}
	generateFile(ctx, exp)
}

// BulkExportObjectToFile export object recursively to file
func BulkExportObjectToFile(ctx *Context, data impexp.ExportData) {
	log.Debugf("DATA %#+v", data)
	exp := impexp.NewExport(data.Info)
	for _, obj := range data.Objects {
		var recursive bool
		if obj.Options != nil {
			recursive = obj.Options.Recursive
		} else {
			recursive = true
		}
		err := exp.Export(obj.ObjectTypeID, obj.ObjectID, recursive, 0)
		if err != nil {
			log.Errorf("Error on create Export Data  for object  %s: %s with: %s", obj.ObjectTypeID, obj.ObjectID, err)
			ctx.JSON(404, err.Error())
			return
		}
		log.Infof("Object type  %s| with ID: %s | exported Successfully %t", obj.ObjectTypeID, obj.ObjectID, recursive)
	}

	generateFile(ctx, exp)
}
