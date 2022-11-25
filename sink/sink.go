package sink

import "xplayer/pkg"

type Sink interface {
	Format() pkg.MetaData
	Output(data []byte) error
	Close() error
}

type Options struct {
	Width  int
	Height int
	Format string
}

type Option func(options *Options)

func Width(w int) Option {
	return func(options *Options) {
		options.Width = w
	}
}
func Height(h int) Option {
	return func(options *Options) {
		options.Height = h
	}
}

func Format(format string) Option {
	return func(options *Options) {
		options.Format = format
	}
}
