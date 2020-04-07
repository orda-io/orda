package restful

import (
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/server/mongodb"
	"net/http"
	"strings"
)

const (
	collectionPath = "/collections/"
)

type Server struct {
	port  int
	mongo *mongodb.RepositoryMongo
}

func NewServer(port int, mongo *mongodb.RepositoryMongo) *Server {
	return &Server{
		port:  port,
		mongo: mongo,
	}
}

func (its *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc(collectionPath, its.collections)

	addr := fmt.Sprintf(":%d", its.port)
	if err := http.ListenAndServe(addr, mux); err != nil {
		return err
	}

	return nil
}

func (its *Server) collections(res http.ResponseWriter, req *http.Request) {
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
