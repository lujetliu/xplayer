package fy

import (
	"image"
	"xplayer/pkg"
	"xplayer/sink"

	"fyne.io/fyne/v2/canvas"
)

type CanvasImage struct {
	option sink.Options
	pix    *image.NRGBA
	image  *canvas.Image
}

func (c *CanvasImage) Format() pkg.MetaData {
	return pkg.MetaData{
		Size:   pkg.Size{Width: c.option.Width, Height: c.option.Height},
		Format: "RGB",
	}
}

func (c *CanvasImage) Close() error {
	return nil
}

func (c *CanvasImage) Output(data []byte) error {
	c.pix.Pix = data
	c.image.Refresh()
	return nil
}

func NewCanvasImageSink(pix *image.NRGBA, i *canvas.Image, opt ...sink.Option) sink.Sink {
	o := sink.Options{}
	for _, option := range opt {
		option(&o)
	}
	return &CanvasImage{option: o, pix: pix, image: i}
}
