package mongodb

import (
	"encoding/json"
	"fmt"
	"github.com/orda-io/orda/pkg/log"
	"net/url"
)

const connectionFormat = "mongodb://%s:%s@%s/"

// Config is a configuration for MongoDB
type Config struct {
	Host     string `json:"MongoHost"`
	OrdaDB   string `json:"OrdaDB"`
	User     string `json:"User"`
	Password string `json:"Password"`
	CertFile string `json:"CertFile"`
	Options  string `json:"Options"`
}

func (its *Config) getConnectionString() string {
	ret := fmt.Sprintf(connectionFormat, its.User, url.QueryEscape(its.Password), its.Host)
	if its.Options == "" {
		return ret
	}
	return ret + "?" + its.Options
}

func (its *Config) String() string {
	clone := struct {
		Config
	}{
		*its,
	}
	clone.Password = ""
	log.Logger.Infof("%v", clone)
	b, _ := json.Marshal(clone)
	return string(b)
}

func (its *Config) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Host     string
		OrdaDB   string
		User     string
		CertFile string
		Options  string
	}{
		Host:     its.Host,
		OrdaDB:   its.OrdaDB,
		User:     its.User,
		CertFile: its.CertFile,
		Options:  its.Options,
	})
}
