package source

import "xplayer/pkg"

type AdbDeviceHandler interface {
	OnNewSource(s Source)
	Closed(s Source)
}

type ServerOption struct {
	MainClass string
	Serial    string
	LocalPort int
	//maxSize   int
	BitRate int
	//crop      string
	//videoSize string
	OverTcp          bool
	Mirror           bool
	AdbDeviceHandler AdbDeviceHandler
}

type Options struct {
	LocalPort  int
	SockName   string
	Serial     string
	ServerPath string
}

type Option func(options *Options)

func LocalPort(port int) Option {
	return func(o *Options) {
		o.LocalPort = port
	}
}

func SockName(s string) Option {
	return func(o *Options) {
		o.SockName = s
	}
}

func Serial(s string) Option {
	return func(o *Options) {
		o.Serial = s
	}
}

func ServerPath(p string) Option {
	return func(o *Options) {
		o.ServerPath = p
	}
}

type Source interface {
	Name() string
	Send(b []byte) error
	Close() error
	//Init() error
	Input(buf []byte) (int, error)
	Format() pkg.MetaData
}

type TcpServer interface {
	AddSourceCallback(handler AdbDeviceHandler)
	Listen()
	Close() error
}
