package tcp

import (
	"go-scrcpy/pkg"
	"go-scrcpy/pkg/log"
	"io"
	"net"
)

const deviceNameLength = 64

// Source holds info about connection
type Source struct {
	name       string
	screenSize pkg.Size
	conn       net.Conn
	s          *server
	close      bool
}

func (c *Source) Name() string {
	return c.name
}

func (c *Source) Format() pkg.MetaData {
	return pkg.MetaData{
		Size:   c.screenSize,
		Format: "h264",
	}
}

func (c *Source) Init() error {
	buf := make([]byte, deviceNameLength+4)
	if _, err := io.ReadFull(c.conn, buf); err != nil {
		return err
	}

	c.name = string(buf[:deviceNameLength])
	c.screenSize.Width = int(buf[deviceNameLength])<<8 | int(buf[deviceNameLength+1])
	c.screenSize.Height = int(buf[deviceNameLength+2])<<8 | int(buf[deviceNameLength+3])
	log.Info("device screen size = %+v, %s", c.screenSize, string(buf))
	return nil
}

func (c *Source) Input(buf []byte) (n int, err error) {
	if c.close {
		return 0, io.EOF
	}
	n, err = c.conn.Read(buf)
	if err != nil {
		log.Error("err %+v", err)
		c.s.adbDeviceHandler.Closed(c)
	}
	return
}

func (c *Source) Send(b []byte) error {
	_, err := c.conn.Write(b)
	if err != nil {
		log.Error("err %+v", err)
		c.s.adbDeviceHandler.Closed(c)
	}
	return err
}

func (c *Source) Conn() net.Conn {
	return c.conn
}

func (c *Source) Close() error {
	c.close = true
	return c.conn.Close()
}
