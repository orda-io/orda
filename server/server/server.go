package server

import (
	"context"
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
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
	httpServer *restful.Server
	ctx        context.Context
	conf       *OrtooServerConfig
	service    *service.OrtooService
	notifier   *notification.Notifier
	Mongo      *mongodb.RepositoryMongo
}

// NewOrtooServer creates a new Ortoo server
func NewOrtooServer(ctx context.Context, conf *OrtooServerConfig) (*OrtooServer, error) {
	mongo, err := mongodb.New(ctx, conf.Mongo)
	if err != nil {
		return nil, log.OrtooError(err)
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
func (its *OrtooServer) Start() error {
	its.mutex.Lock()
	defer its.mutex.Unlock()

	lis, err := net.Listen("tcp", its.conf.OrtooServer)
	if err != nil {
		log.Logger.Fatalf("fail to listen: %v", err)
	}
	its.notifier, err = notification.NewNotifier(its.conf.Notification)
	if err != nil {
		return log.OrtooError(err)
	}

	its.rpcServer = grpc.NewServer()
	if its.service, err = service.NewOrtooService(its.Mongo, its.notifier); err != nil {
		panic("fail to connect MongoDB")
	}

	its.httpServer = restful.NewServer(5000, its.Mongo)

	model.RegisterOrtooServiceServer(its.rpcServer, its.service)
	fmt.Printf("%sStarted at %s\n", banner, time.Now().String())
	go func() {
		if err := its.rpcServer.Serve(lis); err != nil {
			_ = log.OrtooErrorf(err, "fail to serve grpc")
		}

	}()

	go func() {
		if err := its.httpServer.Start(); err != nil {
			_ = log.OrtooErrorf(err, "fail to serve http")
		}
	}()

	log.Logger.Info("successfully start")
	return nil
}

func (its *OrtooServer) Close(graceful bool) {
	its.mutex.Lock()
	defer func() {
		its.mutex.Unlock()
		its.closed = true
	}()

	if graceful {
		log.Logger.Infof("gracefully shutdown server")
		its.rpcServer.GracefulStop()
	} else {
		log.Logger.Infof("Stop server")
		its.rpcServer.Stop()
	}
}

func (its *OrtooServer) HandleSignals() int {
	log.Logger.Infof("ready to handle signals")
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	var sig os.Signal
	select {
	case s := <-signalCh:
		sig = s
	case <-its.closedCh:
		return 0
	}

	log.Logger.Infof("caught signal: %s", sig.String())
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
		log.Logger.Infof("closed successfully")
		return 0
	}
}
