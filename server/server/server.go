package server

import (
	gocontext "context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	"github.com/orda-io/orda/pkg/constants"
	"github.com/orda-io/orda/pkg/context"
	"github.com/orda-io/orda/pkg/errors"
	"github.com/orda-io/orda/pkg/log"
	"github.com/orda-io/orda/pkg/model"
	svrConstant "github.com/orda-io/orda/server/constants"
	"github.com/orda-io/orda/server/mongodb"
	"github.com/orda-io/orda/server/notification"
	"github.com/orda-io/orda/server/service"
	"github.com/orda-io/orda/server/svrcontext"
)

const banner = `       
    .::::                 .::          
  .::    .::              .::          
.::        .::.: .:::     .::   .::    
.::        .:: .::    .:: .:: .::  .:: 
.::        .:: .::   .:   .::.::   .:: 
  .::     .::  .::   .:   .::.::   .:: 
    .::::     .:::    .:: .::  .:: .:::

`

const defaultGracefulTimeout = 10 * time.Second

// OrdaServer is an Orda server
type OrdaServer struct {
	closed     bool
	mutex      sync.Mutex
	closedCh   chan struct{}
	rpcServer  *grpc.Server
	restServer *RestServer
	ctx        context.OrdaContext
	conf       *OrdaServerConfig
	service    *service.OrdaService
	notifier   *notification.Notifier
	Mongo      *mongodb.RepositoryMongo
}

// NewOrdaServer creates a new Orda server
func NewOrdaServer(goCtx gocontext.Context, conf *OrdaServerConfig) (*OrdaServer, errors.OrdaError) {
	host, err := os.Hostname()
	if err != nil {
		return nil, errors.ServerInit.New(log.Logger, err.Error())
	}
	ctx := svrcontext.NewServerContext(goCtx, svrConstant.TagServer).
		UpdateCollection(context.MakeTagInServer(host, conf.RPCServerPort))
	ctx.L().Infof("Config: %#v", conf)
	mongo, oErr := mongodb.New(ctx, &conf.Mongo)
	if oErr != nil {
		return nil, oErr
	}

	return &OrdaServer{
		ctx:      ctx,
		conf:     conf,
		Mongo:    mongo,
		closed:   false,
		closedCh: make(chan struct{}),
	}, nil
}

// Start start the Orda Server
func (its *OrdaServer) Start() errors.OrdaError {
	its.mutex.Lock()
	defer its.mutex.Unlock()

	server := fmt.Sprintf("Orda-Server-%s(%s)", constants.Version, constants.GitCommit)

	var oErr errors.OrdaError
	its.notifier, oErr = notification.NewNotifier(its.ctx, its.conf.Notification, server)
	if oErr != nil {
		return oErr
	}

	lis, err := net.Listen("tcp", its.conf.GetRPCServerAddr())
	if err != nil {
		return errors.ServerInit.New(its.ctx.L(), "cannot listen RPC:"+err.Error())
	}
	its.rpcServer = grpc.NewServer()
	its.service = service.NewOrdaService(its.Mongo, its.notifier)
	model.RegisterOrdaServiceServer(its.rpcServer, its.service)

	go func() {
		its.ctx.L().Infof("open port: tcp://0.0.0.0%s", its.conf.GetRPCServerAddr())
		if err := its.rpcServer.Serve(lis); err != nil {
			_ = errors.ServerInit.New(its.ctx.L(), err.Error())
			panic("fail to serve RPC Server")
		}
	}()

	its.ctx.L().Printf("%s Started at %s %s", server, time.Now().String(), banner)

	its.restServer = New(its.ctx, its.conf, its.Mongo)
	go func() {
		if err := its.restServer.Start(); err != nil {
			_ = errors.ServerInit.New(its.ctx.L(), err.Error())
			panic("fail to serve control server")
		}
	}()

	its.ctx.L().Info("start Orda server successfully")
	return nil
}

func (its *OrdaServer) runRpcGwServer() errors.OrdaError {
	gwMux := runtime.NewServeMux()
	gwOpts := []grpc.DialOption{grpc.WithInsecure()}

	gwErr := model.RegisterOrdaServiceHandlerFromEndpoint(its.ctx, gwMux, its.conf.GetRPCServerAddr(), gwOpts)
	if gwErr != nil {
		return errors.ServerNoResource.New(its.ctx.L(), gwErr.Error())
	}

	mux := http.NewServeMux()
	mux.Handle("/api/", gwMux)

	fs := http.FileServer(http.Dir("./server/swagger-ui"))
	mux.Handle("/swaggerui/", http.StripPrefix("/swaggerui/", fs))

	go func() {
		if err := http.ListenAndServe(":22222", mux); err != nil {
			panic("fail to serve rpc gateway server")
		}
	}()
	return nil
}

// Close closes all the server threads.
func (its *OrdaServer) Close(graceful bool) {
	its.mutex.Lock()
	defer func() {
		its.mutex.Unlock()
		its.closed = true
	}()

	if graceful {
		its.ctx.L().Infof("gracefully shutdown server")
		its.rpcServer.GracefulStop()
	} else {
		its.ctx.L().Infof("Stop server")
		its.rpcServer.Stop()
	}
}

// HandleSignals can handle signals to the server.
func (its *OrdaServer) HandleSignals() int {
	its.ctx.L().Infof("ready to handle signals")
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	var sig os.Signal
	select {
	case s := <-signalCh:
		sig = s
	case <-its.closedCh:
		return 0
	}

	its.ctx.L().Infof("caught signal: %s", sig.String())
	graceful := false
	if sig == syscall.SIGINT || sig == syscall.SIGTERM {
		graceful = true
	}

	gracefulCh := make(chan struct{})
	go func() {
		its.Close(graceful)
		close(gracefulCh)
	}()

	select {
	case <-signalCh:
		return 1
	case <-time.After(defaultGracefulTimeout):
		return 1
	case <-gracefulCh:
		its.ctx.L().Infof("closed successfully")
		return 0
	}
}
