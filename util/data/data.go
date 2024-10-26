package data

import (
	"encoding/json"
	"os"
	"path"

	"github.com/Logiase/MiraiGo-Template/global"
	log "github.com/sirupsen/logrus"
)

type Data map[string]interface{}

func (d Data) SetDefault(key string, default_value interface{}) (result interface{}) {
	v, ok := d[key]
	if ok {
		return v
	} else {
		d[key] = default_value
		return default_value
	}
}

func GetData(name string) Data {
	path := path.Join("data/", name+".json")
	payload := make(Data)
	if !global.PathExists(path) {
		return payload
	}

	content := global.ReadFile(path)
	err := json.Unmarshal(content, &payload)
	if err != nil {
		return make(Data)
	}
	log.Infof("payload: %s", payload)
	return payload
}

func SetData(name string, v any) bool {
	path := path.Join("data/", name+".json")
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Error(err)
		return false
	}
	err = os.WriteFile(path, data, 0222)
	if err != nil {
		log.Error(err)
		return false
	}
	return true
}
