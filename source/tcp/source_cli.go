package tcp

import (
	"fmt"
	"go-scrcpy/internal/source"
	"go-scrcpy/pkg"
	"go-scrcpy/pkg/adb"
	"go-scrcpy/pkg/log"
	"io"
	"net"
	"time"
)

type sourceCli struct {
	options    source.Options
	screenSize pkg.Size
	conn       net.Conn
	close      bool
}

func (c *sourceCli) Name() string {
	return c.options.Serial
}

func (c *sourceCli) Format() pkg.MetaData {
	return pkg.MetaData{
		Size:   c.screenSize,
		Format: "h264",
	}
}

func (c *sourceCli) Init() error {
	buf := make([]byte, deviceNameLength+4)
	if _, err := io.ReadFull(c.conn, buf); err != nil {
		return err
	}

	//c.name = string(buf[:deviceNameLength])
	c.screenSize.Width = int(buf[deviceNameLength])<<8 | int(buf[deviceNameLength+1])
	c.screenSize.Height = int(buf[deviceNameLength+2])<<8 | int(buf[deviceNameLength+3])
	log.Info("device screen size = %+v, %s", c.screenSize, string(buf))
	return nil
}

func (c *sourceCli) Input(buf []byte) (n int, err error) {
	if c.close {
		return 0, io.EOF
	}
	n, err = c.conn.Read(buf)
	if err != nil {
	}
	return
}

func (c *sourceCli) Send(b []byte) error {
	_, err := c.conn.Write(b)
	if err != nil {
		log.Error("err %+v", err)
	}
	return err
}

func (c *sourceCli) Conn() net.Conn {
	return c.conn
}

func (c *sourceCli) Close() error {
	c.close = true
	c.conn.Close()
	return nil
}

func (c *sourceCli) connectToRemote(attempts int, delay, timeout time.Duration) (err error) {
	for attempts > 0 {
		if err = c.connectAndReadByte(timeout); err == nil {
			return
		}
		time.Sleep(delay)
		attempts--
	}

	return
}

func (c *sourceCli) connectAndReadByte(timeout time.Duration) (err error) {
	if c.conn, err = net.DialTimeout(
		"tcp", fmt.Sprintf(":%d", c.options.LocalPort), 5*time.Second); err != nil {
		log.Info("DialTimeout, err = %v", err)
		return
	}
	if timeout > 0 {
		c.conn.SetReadDeadline(time.Now().Add(timeout))
		defer c.conn.SetReadDeadline(time.Time{})
	}

	// 只要 tunnel 建立（adb froward）建连就会成功，
	// 即使此时 device 上的 Device 还没有 listen。
	// 所以这里还要读取一个字节，保证 device 上的 Device 已经开始工作
	buf := make([]byte, 1)
	_, err = io.ReadFull(c.conn, buf)
	return
}

func NewSourceClient(opt ...source.Option) (source.Source, error) {
	o := source.Options{}
	for _, option := range opt {
		option(&o)
	}
	c := &sourceCli{
		options: o,
		conn:    nil,
		close:   false,
	}
	err := c.enableTunnelForward()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := c.disableTunnelForward(); err != nil {
			log.Error("disableTunnelForward fail, err = %v", err)
		}
	}()

	if process, err := adb.ExecAsync(c.options.Serial, "shell",
		fmt.Sprintf("CLASSPATH=%s", c.options.ServerPath),
		"app_process",
		"/",
		"com.genymobile.scrcpy.Server",
		fmt.Sprintf("%d", 8000000),
		"true",
		"false"); err != nil {
		return nil, err
	} else {
		log.Info("process %v", process.Process.Pid)
	}
	if err = c.connectToRemote(100, 100*time.Millisecond, 10*time.Second); err != nil {
		return nil, err
	}
	c.Init()
	return c, nil
}

func (c *sourceCli) enableTunnelForward() error {
	return adb.Forward(c.options.Serial, c.options.LocalPort, c.options.SockName)
}

func (c *sourceCli) disableTunnelForward() error {
	return adb.ForwardRemove(c.options.Serial, c.options.LocalPort)
}
