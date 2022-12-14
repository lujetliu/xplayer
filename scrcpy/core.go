package scrcpy

import "C"
import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"io"
	"log"
	"net"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"xplayer/h264"

	"github.com/xmsociety/adbutils"
	"gocv.io/x/gocv"
)

func substr(s string, pos, length int) string {
	runes := []rune(s)
	l := pos + length
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[pos:l])
}

func getParentDirectory(dirctory string) string {
	return substr(dirctory, 0, strings.LastIndex(dirctory, "/"))
}

func getCurrentFile() string {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		panic(errors.New("Can not get current file info"))
	}
	return getParentDirectory(file)
}

type resolution struct {
	W uint16
	H uint16
}

type Client struct {
	Device                adbutils.AdbDevice
	MaxWith               int
	Bitrate               int //8000000
	MaxFps                int
	Alive                 bool
	Flip                  bool
	BlockFrame            bool
	StayAwake             bool
	LockScreenOrientation int
	ConnectionTimeout     int    //3000
	EncoderName           string // "-"
	serverStream          adbutils.AdbConnection
	videoSocket           net.Conn
	controlSocket         net.Conn
	VideoSender           chan image.Image
	ErrReceiver           chan error
	Resolution            resolution
	Control               ControlSender
	Ctx                   context.Context
	Lock                  sync.Mutex
}

type ReadDeadLiner interface {
	SetReadDeadline(t time.Time) error
}

// SetReadDeadlineOnCancel graceful close socket
func (client *Client) SetReadDeadlineOnCancel() {
	// go func() {
	// 	<-client.Ctx.Done()
	// 	now := time.Now()
	// 	_ = client.videoSocket.SetReadDeadline(now)
	// 	_ = client.controlSocket.SetReadDeadline(now)
	// 	_ = client.serverStream.Conn.SetReadDeadline(now)
	// 	client.Stop()
	// }()
}

func readFully(conn net.Conn, n int) ([]byte, error) {
	t := 0
	var err error
	buffer := make([]byte, n)
	result := bytes.NewBuffer(nil)
	for t < n {
		length, err := conn.Read(buffer[0:n])
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Println(err.Error())
			}
			return result.Bytes(), err
		}
		result.Write(buffer[0:length])
		t += length

	}
	return result.Bytes(), err
}

func (client *Client) deployServer() bool {
	jarName := "scrcpy-server.jar"
	currentPath := getCurrentFile()
	src, _ := filepath.Abs(path.Join(currentPath, jarName))
	client.Device.Push(src, fmt.Sprintf("/data/local/tmp/%v", jarName))
	stayAwake := "false"
	if client.StayAwake {
		stayAwake = "true"
	}
	// CLASSPATH=/data/local/tmp/scrcpy-server.jar   app_process / com.genymobile.scrcpy.Server 1.20  info 0 100000 0 -1 true - false ture 0 false true - - false
	cmd := []string{
		fmt.Sprintf("CLASSPATH=/data/local/tmp/%v", jarName),
		"app_process",
		"/",
		"com.genymobile.scrcpy.Server",
		"1.20",                       // Scrcpy server version
		"debug",                      // Log level: info, verbose...
		strconv.Itoa(client.MaxWith), // Max screen width (long side)
		strconv.Itoa(client.Bitrate), // Bitrate of video
		strconv.Itoa(client.MaxFps),  // Max frame per second
		strconv.Itoa(LockScreenOrientationUnlocked), // Lock screen orientation: LOCK_SCREEN_ORIENTATION
		"true",    // Tunnel forward
		"-",       // Crop screen
		"false",   // Send frame rate to client
		"true",    // Control enabled
		"0",       // Display id
		"false",   // Show touches
		stayAwake, // Stay awake
		"-",       // Codec (video encoding) options
		"-",       // Encoder name
		"false",   // Power off screen after server closed
	}
	serverStream := client.Device.Shell(strings.Join(cmd, " "), true, 3)
	client.serverStream = *serverStream.(*adbutils.AdbConnection)
	// res := client.serverStream.ReadString(100) wait serverStream ready
	client.serverStream.ReadString(100) // wait serverStream ready
	time.Sleep(time.Millisecond * 500)  // adjust when poor network
	fmt.Println("deploy success")
	return true
}

func (client *Client) initServerConnection() bool {
	if client.ConnectionTimeout == 0 {
		client.ConnectionTimeout = 3000
	}
	for i := 0; i < client.ConnectionTimeout; i += 100 {
		client.videoSocket = client.Device.CreateConnection(adbutils.LOCALABSTRACT, "scrcpy")
		if client.videoSocket != nil {
			break
		} else {
			time.Sleep(time.Millisecond * 100)
		}
	}
	if client.videoSocket == nil {
		client.ErrReceiver <- errors.New("Failed to connect scrcpy-server after 3 seconds please wait a moment to retry!")
		log.Printf("%v: Failed to connect scrcpy-server after 3 seconds", client.Device.Serial)
		return false
	}
	buf, _ := readFully(client.videoSocket, 1)
	if buf == nil || len(buf) == 0 || buf[0] != []byte("\x00")[0] {
		client.ErrReceiver <- errors.New("Did not receive Dummy Byte! please wait a moment to retry!")
		log.Printf("%v: Did not receive Dummy Byte!", client.Device.Serial)
		return false
	}
	client.controlSocket = client.Device.CreateConnection(adbutils.LOCALABSTRACT, "scrcpy")
	nameBuf, _ := readFully(client.videoSocket, 64)
	if nameBuf == nil || len(nameBuf) == 0 || strings.TrimSuffix(string(nameBuf), "\x00") == "" {
		client.ErrReceiver <- errors.New(fmt.Sprintf("Did not receive Device Name! err: %v please wait a moment to retry!", nameBuf))
		log.Printf("%v: Did not receive Device Name! err", client.Device.Serial)
		return false
	}
	resBuf, _ := readFully(client.videoSocket, 4)
	r := bytes.NewReader(resBuf)

	resolutionTmp := resolution{}
	if err := binary.Read(r, binary.BigEndian, &resolutionTmp); err != nil {
		client.ErrReceiver <- errors.New(fmt.Sprintf("binary.Read failed: %v, please wait a moment to retry!", err))
		log.Printf("%v binary.Read failed:", err.Error())
		return false
	}
	client.Resolution = resolutionTmp
	client.Control = ControlSender{
		ControlConn: client.controlSocket,
		W:           int(client.Resolution.W),
		H:           int(client.Resolution.H),
		Lock:        sync.Mutex{},
	}
	log.Printf("%v: client init W, H: %v", client.Device.Serial, client.Resolution)
	return true
}

func (client *Client) Start() {
	defer func() {
		log.Printf("%v: goroutine client.start quit!", client.Device.Serial)
		client.Stop()
	}()
	client.Alive = true
	if !client.deployServer() {
		return
	}
	if !client.initServerConnection() {
		return
	}
	client.streamLoop()

}

func (client *Client) Stop() {
	client.Alive = false
	if client.serverStream.Conn != nil {
		client.serverStream.Close()
	}
	if client.controlSocket != nil {
		client.controlSocket.Close()
	}
	if client.videoSocket != nil {
		client.videoSocket.Close()
	}
}

func (client *Client) streamLoop() {
	// thanks https://github.com/mike1808/h264decoder totaly what I want
	codec, err := h264.NewDecoder(h264.PixelFormatBGR)
	defer func() {
		codec.Close()
		log.Printf("%v: function streamLoop quit!", client.Device.Serial)
	}()
	if err != nil {
		log.Println(err.Error())
		return
	}
	client.SetReadDeadlineOnCancel()
	// TODO need fix -> could not determine kind of name for C.sws_addVec
	for client.Alive {
		buf, err := readFully(client.videoSocket, 1024)
		if err != nil {
			return
		}
		frames, err := codec.Decode(buf)
		if err != nil {
			log.Println(err.Error())
			return
		}
		for _, frame := range frames {
			client.Resolution = resolution{W: uint16(frame.Width), H: uint16(frame.Height)}
			client.Control.W = frame.Width
			client.Control.H = frame.Height
			imgCv, _ := gocv.NewMatFromBytes(frame.Height, frame.Width, gocv.MatTypeCV8UC3, frame.Data)
			if imgCv.Empty() {
				continue
			}
			imageRGB, err := imgCv.ToImage()
			if err != nil {
				log.Println(err.Error())
				continue
			}
			if !client.Alive {
				log.Printf("%v: client not alive stop send imageRGB!", client.Device.Serial)
				return
			}
			client.VideoSender <- imageRGB
		}
		time.Sleep(time.Microsecond * 100)

	}
}
