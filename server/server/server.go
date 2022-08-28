package server

import (
	gocontext "context"
	"fmt"
	"github.com/orda-io/orda/client/pkg/constants"
	"github.com/orda-io/orda/client/pkg/context"
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/log"
	"github.com/orda-io/orda/client/pkg/model"
	"github.com/orda-io/orda/server/managers"
	"google.golang.org/grpc/reflection"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"

	svrConstant "github.com/orda-io/orda/server/constants"
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
	conf       *managers.OrdaServerConfig
	service    *service.OrdaService
	managers   *managers.Managers
}

// NewOrdaServer creates a new Orda server
func NewOrdaServer(goCtx gocontext.Context, conf *managers.OrdaServerConfig) (*OrdaServer, errors.OrdaError) {
	host, err := os.Hostname()
	if err != nil {
		return nil, errors.ServerInit.New(log.Logger, err.Error())
	}
	ctx := svrcontext.NewServerContext(goCtx, svrConstant.TagServer).
		UpdateCollection(context.MakeTagInServer(host, conf.RPCServerPort))
	ctx.L().Infof("Config: %v", conf)
	return &OrdaServer{
		ctx:      ctx,
		conf:     conf,
		closed:   false,
		closedCh: make(chan struct{}),
	}, nil
}

// Start starts the Orda Server
func (its *OrdaServer) Start() errors.OrdaError {
	its.mutex.Lock()
	defer its.mutex.Unlock()

	var oErr errors.OrdaError
	if its.managers, oErr = managers.New(its.ctx, its.conf); oErr != nil {
		return oErr
	}

	lis, err := net.Listen("tcp", its.conf.GetRPCServerAddr())
	if err != nil {
		return errors.ServerInit.New(its.ctx.L(), "fail to listen RPC:"+err.Error())
	}
	its.rpcServer = grpc.NewServer()
	reflection.Register(its.rpcServer)
	its.service = service.NewOrdaService(its.managers)
	model.RegisterOrdaServiceServer(its.rpcServer, its.service)

	go func() {
		its.ctx.L().Infof("open port: tcp://0.0.0.0%s", its.conf.GetRPCServerAddr())
		if err := its.rpcServer.Serve(lis); err != nil {
			_ = errors.ServerInit.New(its.ctx.L(), err.Error())
			panic("fail to serve RPC Server")
		}
	}()

	its.restServer = NewRestServer(its.ctx, its.conf, its.managers)
	go func() {
		if err := its.restServer.Start(); err != nil {
			_ = errors.ServerInit.New(its.ctx.L(), err.Error())
			panic("fail to serve control server")
		}
	}()
	server := fmt.Sprintf("Orda-Server-%s (%s)", constants.Version, constants.BuildInfo)
	its.ctx.L().Infof("%s Started at %s %s", server, time.Now().String(), banner)
	its.ctx.L().Info("start Orda server successfully")
	return nil
}

// Close closes all the server threads.
func (its *OrdaServer) Close(graceful bool) {
	its.mutex.Lock()
	defer func() {
		its.managers.Close(its.ctx)
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
