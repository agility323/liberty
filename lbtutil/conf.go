/*
By Thomas Wade, 2022.02.25
*/
package lbtutil

import (
	"flag"
	"errors"
	"encoding/json"
	"reflect"
	"os"
	"io/ioutil"

	"github.com/mohae/deepcopy"
)

var (
	confData = map[string]interface{}{}
)

type flagSetData struct {
	Pdata *interface{}
}

func (v flagSetData) String() string {
	if v.Pdata == nil { return "" }
	b, _ := json.Marshal(*v.Pdata)
	return string(b)
}

func (v flagSetData) Set(s string) error {
	if v.Pdata == nil { return errors.New("load conf fail: flagSetData.Pdata is nil") }
	// remain string type
	if reflect.TypeOf(*v.Pdata) == reflect.TypeOf("") {
		*v.Pdata = s
		return nil
	}
	// parse with json
	err := json.Unmarshal([]byte(s), v.Pdata)
	return err
}

func LoadConfFromCmdLine(defs map[string]interface{}, args []string, pconf interface{}) {
	confs := make([]map[string]interface{}, 0, 10)
	// load cmd line conf
	parser := flag.NewFlagSet("conf", flag.ExitOnError)
	ptrs := make(map[string]*interface{}, len(defs))
	for name, val := range defs {
		valcp := deepcopy.Copy(val)
		ptrs[name] = &valcp
		fval := flagSetData{Pdata: &valcp}
		parser.Var(fval, name, "")
	}
	confFile := ""
	parser.StringVar(&confFile, "conf", "", "main conf file")
	parser.Parse(args)
	cmdconf := make(map[string]interface{}, len(ptrs))
	for k, pv := range ptrs {
		cmdconf[k] = *pv
	}
	confs = append(confs, cmdconf)
	// load file conf
	log.Info("load conf file %s", confFile)
	if confFile != "" {
		fconf, err := loadConfFromFile(confFile)
		if err == nil {
			confs = append(confs, fconf)
		} else {
			log.Error("invalid conf file %s %s", confFile, err.Error())
		}
	}
	// merge conf
	confData := make(map[string]interface{})
	for _, conf := range confs {
		for k, v := range conf { confData[k] = v }
	}
	// show conf
	for k, v := range confData { log.Info("conf entry %s %v", k, v) }
	// writo to conf
	b, _ := json.Marshal(confData)
	if err := json.Unmarshal(b, pconf); err != nil {
		log.Error("load conf fail %s", err.Error())
	}
}

func loadConfFromFile(fn string) (map[string]interface{}, error) {
	conf := make(map[string]interface{})
	f, err := os.Open(fn)
	if err != nil { return conf, err }
	defer f.Close()
	fdata, err := ioutil.ReadAll(f)
	if err != nil { return conf, err }
	err = json.Unmarshal(fdata, &conf)
	if err != nil { return conf, err }
	return conf, nil
}
