package server

import (
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/orda-io/orda/pkg/context"
	"github.com/orda-io/orda/pkg/errors"
	"github.com/orda-io/orda/pkg/model"
	"github.com/orda-io/orda/server/mongodb"
	"github.com/swaggo/http-swagger"
	"github.com/swaggo/swag"
	"google.golang.org/grpc"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	apiGrpcGw        = "/api/"
	apiGrpcGwSwagger = "/swagger/"
	apiCollections   = "/collections/"
	swaggerJson      = "./proto/orda.grpc.swagger.json"
)

// RestServer is a control server to set up Orda system.
type RestServer struct {
	ctx   context.OrdaContext
	conf  *OrdaServerConfig
	mongo *mongodb.RepositoryMongo
}

// New creates a control server.
func New(ctx context.OrdaContext, conf *OrdaServerConfig, mongo *mongodb.RepositoryMongo) *RestServer {

	return &RestServer{
		ctx:   ctx,
		conf:  conf,
		mongo: mongo,
	}
}

// Start starts a RestServer
func (its *RestServer) Start() errors.OrdaError {

	mux := http.NewServeMux()

	if err := its.initGrpcGatewayServer(mux); err != nil {
		return err
	}

	if err := http.ListenAndServe(its.conf.GetRestfulAddr(), its.allowCors(mux)); err != nil {
		return errors.ServerInit.New(its.ctx.L(), err)
	}
	return nil
}

func (its *RestServer) allowCors(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				headers := []string{"Content-Type", "Accept"}
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
				methods := []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodDelete}
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

type swaggerDoc struct{}

func (its *swaggerDoc) ReadDoc() string {
	bytes, err := ioutil.ReadFile(swaggerJson)
	if err != nil {
		panic("")
	}
	return string(bytes)
}

func (its *RestServer) initGrpcGatewayServer(mux *http.ServeMux) errors.OrdaError {
	gwMux := runtime.NewServeMux()
	gwOpts := []grpc.DialOption{grpc.WithInsecure()}

	if gwErr := model.RegisterOrdaServiceHandlerFromEndpoint(its.ctx, gwMux, its.conf.GetRPCServerAddr(), gwOpts); gwErr != nil {
		return errors.ServerInit.New(its.ctx.L(), gwErr.Error())
	}

	mux.Handle(apiGrpcGw, gwMux)
	its.ctx.L().Infof("open port: http://localhost%s%s",
		its.conf.GetRestfulAddr(), apiGrpcGw)
	swag.Register(swag.Name, &swaggerDoc{})

	swaggerHandler := httpSwagger.Handler(httpSwagger.URL("./doc.json"))

	mux.Handle(apiGrpcGwSwagger, swaggerHandler)
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
