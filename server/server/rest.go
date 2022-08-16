package server

import (
	"fmt"
	"github.com/orda-io/orda/client/pkg/context"
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/model"
	"github.com/orda-io/orda/server/managers"
	"google.golang.org/grpc/credentials/insecure"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/swaggo/swag"
	"google.golang.org/grpc"

	"github.com/orda-io/orda/server/mongodb"
)

const (
	apiGrpcGw        = "/api/"
	apiGrpcGwSwagger = "/swagger/"
	apiCollections   = "/collections/"
)

// RestServer is a control server to set up Orda system.
type RestServer struct {
	ctx      context.OrdaContext
	conf     *managers.OrdaServerConfig
	managers *managers.Managers
}

// NewRestServer creates a control server.
func NewRestServer(ctx context.OrdaContext, conf *managers.OrdaServerConfig, clients *managers.Managers) *RestServer {

	return &RestServer{
		ctx:      ctx,
		conf:     conf,
		managers: clients,
	}
}

// Start starts a RestServer
func (its *RestServer) Start() errors.OrdaError {

	mux := http.NewServeMux()

	mux.HandleFunc("/", its.echo)

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
				w.Header().Set("Access-Control-Allow-Headers", r.Header.Get("Access-Control-Request-Headers"))
				methods := []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodDelete}
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

type swaggerDoc struct {
	jsonDoc string
}

func (its *swaggerDoc) init(conf *managers.OrdaServerConfig) {
	bytes, err := ioutil.ReadFile(conf.SwaggerJSON)
	if err != nil {
		panic(err.Error())
	}
	its.jsonDoc = string(bytes)
	if conf.SwaggerBasePath != "" {
		its.jsonDoc = strings.ReplaceAll(its.jsonDoc, "\"/api/", "\"/"+strings.Trim(conf.SwaggerBasePath, "/")+"/api/")
	}
}

func (its *swaggerDoc) ReadDoc() string {
	return its.jsonDoc
}

func (its *RestServer) initGrpcGatewayServer(mux *http.ServeMux) errors.OrdaError {
	gwMux := runtime.NewServeMux()

	// register grpc proxy
	gwOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if gwErr := model.RegisterOrdaServiceHandlerFromEndpoint(its.ctx, gwMux, its.conf.GetRPCServerAddr(), gwOpts); gwErr != nil {
		return errors.ServerInit.New(its.ctx.L(), gwErr.Error())
	}
	mux.Handle(apiGrpcGw, gwMux)
	its.ctx.L().Infof("open port: http://localhost%s%s",
		its.conf.GetRestfulAddr(), apiGrpcGw)

	// register swagger for grpc
	swaggerJSON := &swaggerDoc{}
	swaggerJSON.init(its.conf)
	swag.Register(swag.Name, swaggerJSON)
	swaggerHandler := httpSwagger.Handler(httpSwagger.URL("./doc.json"))
	mux.Handle(apiGrpcGwSwagger, swaggerHandler)
	its.ctx.L().Infof("open port: http://localhost%s%s", its.conf.GetRestfulAddr(), apiGrpcGwSwagger)

	return nil
}

func (its *RestServer) echo(w http.ResponseWriter, r *http.Request) {
	its.ctx.L().Infof("ignored request: %v  %v%v", r.Method, r.Host, r.URL)
}

func (its *RestServer) createCollections(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPut:
		collectionName := strings.TrimPrefix(req.URL.Path, apiCollections)
		num, err := mongodb.MakeCollection(its.ctx, its.managers.Mongo, collectionName)
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
