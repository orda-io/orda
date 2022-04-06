package mongodb

import (
	"fmt"
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
