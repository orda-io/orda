package server

import (
	gocontext "context"
	"fmt"
	"github.com/knowhunger/ortoo/pkg/context"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/utils"
	"github.com/knowhunger/ortoo/pkg/version"
	"github.com/knowhunger/ortoo/server/mongodb"
	"github.com/knowhunger/ortoo/server/notification"
	"github.com/knowhunger/ortoo/server/restful"
	"github.com/knowhunger/ortoo/server/service"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const banner = `       
     _/_/                _/                         
  _/    _/  _/  _/_/  _/_/_/_/    _/_/      _/_/    
 _/    _/  _/_/        _/      _/    _/  _/    _/   
_/    _/  _/          _/      _/    _/  _/    _/    
 _/_/    _/            _/_/    _/_/      _/_/

`

const defaultGracefulTimeout = 10 * time.Second

// OrtooServer is an Ortoo server
type OrtooServer struct {
	closed     bool
	mutex      sync.Mutex
	closedCh   chan struct{}
	rpcServer  *grpc.Server
	ctrlServer *restful.ControlServer
	ctx        context.OrtooContext
	conf       *OrtooServerConfig
	service    *service.OrtooService
	notifier   *notification.Notifier
	Mongo      *mongodb.RepositoryMongo
}

// NewOrtooServer creates a new Ortoo server
func NewOrtooServer(goCtx gocontext.Context, conf *OrtooServerConfig) (*OrtooServer, errors.OrtooError) {
	host, err := os.Hostname()
	if err != nil {
		return nil, errors.ServerInit.New(log.Logger, err.Error())
	}
	subLevel := fmt.Sprintf("%0.10s:%d", utils.TrimLong(host, 15), conf.RPCServerPort)
	ctx := context.NewWithTag(goCtx, context.SERVER, subLevel)
	ctx.L().Infof("Config: %#v", conf)
	mongo, oErr := mongodb.New(ctx, &conf.Mongo)
	if oErr != nil {
		return nil, oErr
	}

	return &OrtooServer{
		ctx:      ctx,
		conf:     conf,
		Mongo:    mongo,
		closed:   false,
		closedCh: make(chan struct{}),
	}, nil
}

// Start start the Ortoo Server
func (its *OrtooServer) Start() errors.OrtooError {
	its.mutex.Lock()
	defer its.mutex.Unlock()

	lis, err := net.Listen("tcp", its.conf.getRPCServerAddr())
	if err != nil {
		return errors.ServerInit.New(its.ctx.L(), "cannot listen RPC:"+err.Error())
	}
	var oErr errors.OrtooError
	its.notifier, oErr = notification.NewNotifier(its.ctx, its.conf.Notification)
	if oErr != nil {
		return oErr
	}

	its.rpcServer = grpc.NewServer()
	its.service = service.NewOrtooService(its.Mongo, its.notifier)
	model.RegisterOrtooServiceServer(its.rpcServer, its.service)

	its.ctrlServer = restful.New(its.ctx, its.conf.RestfulPort, its.Mongo)
	its.ctx.L().Printf("%s(%s) Started at %s %s",
		version.Version,
		version.GitCommit,
		time.Now().String(),
		banner)
	go func() {
		if err := its.rpcServer.Serve(lis); err != nil {
			_ = errors.ServerInit.New(its.ctx.L(), err.Error())
			panic("fail to serve RPC Server")
		}

	}()

	go func() {
		if err := its.ctrlServer.Start(); err != nil {
			_ = errors.ServerInit.New(its.ctx.L(), err.Error())
			panic("fail to serve control server")
		}
	}()

	its.ctx.L().Info("successfully start Ortoo server")
	return nil
}

// Close closes all the server threads.
func (its *OrtooServer) Close(graceful bool) {
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
func (its *OrtooServer) HandleSignals() int {
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
