package rdns

import (
	"context"
	"crypto/tls"
	"expvar"
	"net"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// Read/Write timeout in the admin server
const adminServerTimeout = 10 * time.Second

// AdminListener is a DNS listener/server for admin services.
type AdminListener struct {
	httpServer *http.Server

	id   string
	addr string
	opt  AdminListenerOptions

	mux *http.ServeMux
}

var _ Listener = &AdminListener{}

// AdminListenerOptions contains options used by the admin service.
type AdminListenerOptions struct {
	ListenOptions
	TLSConfig *tls.Config
}

// NewAdminListener returns an instance of an admin service listener.
func NewAdminListener(id, addr string, opt AdminListenerOptions) (*AdminListener, error) {
	l := &AdminListener{
		id:   id,
		addr: addr,
		opt:  opt,
		mux:  http.NewServeMux(),
	}
	// Serve metrics.
	l.mux.Handle("/routedns/vars", expvar.Handler())
	return l, nil
}

// Start the admin server.
func (s *AdminListener) Start() error {
	Log.WithFields(logrus.Fields{"id": s.id, "protocol": "tcp", "addr": s.addr}).Info("starting listener")
	return s.startTCP()
}

// Start the admin server with TCP transport.
func (s *AdminListener) startTCP() error {
	s.httpServer = &http.Server{
		Addr:         s.addr,
		TLSConfig:    s.opt.TLSConfig,
		Handler:      s.mux,
		ReadTimeout:  adminServerTimeout,
		WriteTimeout: adminServerTimeout,
	}

	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer ln.Close()
	return s.httpServer.ServeTLS(ln, "", "")
}

// Stop the server.
func (s *AdminListener) Stop() error {
	Log.WithFields(logrus.Fields{"id": s.id, "protocol": "tcp", "addr": s.addr}).Info("stopping listener")
	return s.httpServer.Shutdown(context.Background())
}

func (s *AdminListener) String() string {
	return s.id
}
