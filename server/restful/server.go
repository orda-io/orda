package restful

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/server/mongodb"
	"net/http"
	"strings"
)

const (
	collectionPath = "/collections/"
)

// ControlServer is a control server to set up Ortoo system.
type ControlServer struct {
	port  int
	mongo *mongodb.RepositoryMongo
}

// New creates a control server.
func New(port int, mongo *mongodb.RepositoryMongo) *ControlServer {
	return &ControlServer{
		port:  port,
		mongo: mongo,
	}
}

// Start starts a ControlServer
func (its *ControlServer) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc(collectionPath, its.createCollections)

	addr := fmt.Sprintf(":%d", its.port)
	if err := http.ListenAndServe(addr, mux); err != nil {
		return err
	}

	return nil
}

func (its *ControlServer) createCollections(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		collectionName := strings.TrimPrefix(req.URL.Path, collectionPath)
		num, err := mongodb.MakeCollection(its.mongo, collectionName)
		var msg string
		if err != nil {
			msg = fmt.Sprintf("Fail to create collection '%s'", collectionName)
		} else {
			msg = fmt.Sprintf("Created collection '%s(%d)'", collectionName, num)
		}
		_, err = res.Write([]byte(msg))
		if err != nil {
			log.Logger.Error("fail to response for %s", req.URL.Path)
			return
		}
		log.Logger.Infof(msg)
	}
}
