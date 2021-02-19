package server

import (
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/knowhunger/ortoo/pkg/context"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/server/mongodb"
	"google.golang.org/grpc"
	"net/http"
	"strings"
)

const (
	apiGrpcGw        = "/api/"
	apiGrpcGwSwagger = "/swagger/"
	apiCollections   = "/collections/"
)

// RestServer is a control server to set up Ortoo system.
type RestServer struct {
	ctx   context.OrtooContext
	conf  *OrtooServerConfig
	mongo *mongodb.RepositoryMongo
}

// New creates a control server.
func New(ctx context.OrtooContext, conf *OrtooServerConfig, mongo *mongodb.RepositoryMongo) *RestServer {

	return &RestServer{
		ctx:   ctx,
		conf:  conf,
		mongo: mongo,
	}
}

// Start starts a RestServer
func (its *RestServer) Start() errors.OrtooError {

	mux := http.NewServeMux()

	if err := its.initGrpcGatewayServer(mux); err != nil {
		return err
	}

	if err := http.ListenAndServe(its.conf.GetRestfulAddr(), mux); err != nil {
		return errors.ServerInit.New(its.ctx.L(), err)
	}
	return nil
}

func (its *RestServer) initGrpcGatewayServer(mux *http.ServeMux) errors.OrtooError {
	gwMux := runtime.NewServeMux()
	gwOpts := []grpc.DialOption{grpc.WithInsecure()}

	if gwErr := model.RegisterOrtooServiceHandlerFromEndpoint(its.ctx, gwMux, its.conf.GetRPCServerAddr(), gwOpts); gwErr != nil {
		return errors.ServerInit.New(its.ctx.L(), gwErr.Error())
	}

	mux.Handle(apiGrpcGw, gwMux)
	its.ctx.L().Infof("open port: http://localhost%s%s",
		its.conf.GetRestfulAddr(), apiGrpcGw)
	fs := http.FileServer(http.Dir("./server/swagger-ui"))
	mux.Handle(apiGrpcGwSwagger, http.StripPrefix(apiGrpcGwSwagger, fs))
	its.ctx.L().Infof("open port: http://localhost%s%s",
		its.conf.GetRestfulAddr(), apiGrpcGwSwagger)

	return nil
}

func (its *RestServer) createCollections(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPut:
		collectionName := strings.TrimPrefix(req.URL.Path, apiCollections)
		num, err := mongodb.MakeCollection(its.ctx, its.mongo, collectionName)
		var msg string
		if err != nil {
			msg = fmt.Sprintf("Fail to create collection '%s'", collectionName)
		} else {
			msg = fmt.Sprintf("Created collection '%s(%d)'", collectionName, num)
		}
		_, err2 := res.Write([]byte(msg))
		if err2 != nil {
			_ = errors.ServerInit.New(its.ctx.L(), err2.Error())
			return
		}
		its.ctx.L().Infof(msg)
	}
}
