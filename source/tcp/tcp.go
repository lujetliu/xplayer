package tcp

import (
	"crypto/tls"
	"fmt"
	"go-scrcpy/internal/source"
	"go-scrcpy/pkg/log"
	"net"
)

type server struct {
	address          string
	config           *tls.Config
	adbDeviceHandler source.AdbDeviceHandler
	listener         net.Listener
}

// Listen starts network server
func (s *server) Listen() {
	var listener net.Listener
	var err error
	if s.config == nil {
		listener, err = net.Listen("tcp", s.address)
	} else {
		listener, err = tls.Listen("tcp", s.address, s.config)
	}
	if err != nil {
		log.Error("Error starting TCP server.")
	}
	s.listener = listener
	defer listener.Close()
	for {
		conn, _ := listener.Accept()
		client := &Source{
			conn: conn,
			s:    s,
		}
		if err := client.Init(); err != nil {
			log.Error("client init fail, err %v", err)
			_ = conn.Close()
			continue
		}
		s.adbDeviceHandler.OnNewSource(client)
	}
}

func (s *server) AddSourceCallback(handler source.AdbDeviceHandler) {
	s.adbDeviceHandler = handler
}

func (s *server) Close() error {
	return s.listener.Close()
}

func NewServer(opt source.ServerOption) (source.TcpServer, error) {
	log.Info("Creating server with address %+v", opt)
	server := &server{
		address:          fmt.Sprintf("%s:%d", "localhost", opt.LocalPort),
		config:           nil,
		adbDeviceHandler: &defaultHandler{},
	}

	return server, nil
}

func NewWithTLS(address string, certFile string, keyFile string) *server {
	log.Info("Creating server with address", address)
	cert, _ := tls.LoadX509KeyPair(certFile, keyFile)
	config := tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	server := &server{
		address: address,
		config:  &config,
	}

	return server
}

type defaultHandler struct {
}

func (d *defaultHandler) OnNewSource(c source.Source) {
	log.Info("OnNewSource %s", c.Name())
}

func (d *defaultHandler) Closed(c source.Source) {

}
