package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"log"
	"net/http"
	"os"
	"xplayer/scrcpy"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	CancelMap      = make(map[string]context.CancelFunc)
	VideoImageChan = make(chan image.Image, 10000)
)

type Service struct {
	Client      *scrcpy.Client
	Config      string
	ErrReceiver chan error
}

var GlobalSvc *Service

func NewService(client *scrcpy.Client, ErrReceiver chan error, config string) *Service {
	svc := &Service{Client: client, Config: config, ErrReceiver: ErrReceiver}
	GlobalSvc = svc
	return svc
}

func (s *Service) Start() {
	defer func() {
		log.Printf("[run] %v: function service.client.start quit！", s.Client.Device.Serial)
	}()
	go s.Client.Start()
	go s.handler()
}

var frameCount uint = 0

func (s *Service) handler() {
	defer func() {
		log.Printf("[run] %v: goroutine handler quit！", s.Client.Device.Serial)
	}()
	for {
		select {
		case frame := <-s.Client.VideoSender:
			if !s.Client.Alive {
				return
			}
			// TODO AI
			frameCount++
			// _ = frame
			log.Printf("[run] %v: get Frame -> %v", s.Client.Device.Serial, frame.Bounds())
			// send error when logic
			// s.ErrReceiver <- errors.New("fuck you")
			// ConvFile(frame)
			// data := ConvImg(frame)
			// rabbitmq.Publish("video-image", data)
		case err := <-s.Client.ErrReceiver:
			// receive frame error
			log.Printf("[run] %v: receive Err -> %v", s.Client.Device.Serial, err.Error())
			s.ErrReceiver <- err
			return
		}
	}
}

func (s *Service) Stop() {
	cancelFunc := CancelMap[s.Client.Device.Serial]
	cancelFunc()
}

func ConvImg(img image.Image) []byte {
	b := bytes.NewBuffer(nil)
	err := png.Encode(b, img)
	if err != nil {
		return nil
	}
	return b.Bytes()
}

func ConvFile(img image.Image) {
	fn := fmt.Sprintf("screen_%v.png", frameCount)
	f, err := os.Create(fn)
	if err != nil {
		return
	}
	defer f.Close()

	err = png.Encode(f, img)
	if err != nil {
		return
	}
	return
}

func ServerH264(c *gin.Context) {
	// c.Header("Content-Type", "video/mp2ts")
	// c.Header("Content-Length", strconv.Itoa(len(item.Data)))
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	defer conn.Close()

	fmt.Println("/////////////////////// connet")

	go func() {
		// log.Printf("Received: %s, MessageType: %v", message, messageType)
		for img := range GlobalSvc.Client.VideoSender {
			imgData := ConvImg(img)
			// res := base64.URLEncoding.EncodeToString(imgData)
			err = conn.WriteMessage(websocket.BinaryMessage, imgData)
			if err != nil {
				log.Println("Error during message writing:", err)
				continue
				// break
			}
		}
	}()

	for {
		// 接收前端控制事件
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error during message reading:", err)
			break
		}
		fmt.Println("Receive websocket message: ", messageType, string(message))

		if messageType != websocket.TextMessage {
			continue
		}

		// 解析事件消息
		var event scrcpy.Event
		err = json.Unmarshal(message, &event)
		if err != nil {
			fmt.Println("unmarshal", err, string(message))
		}

		// 非事件消息
		if event.IsEvent != 1 {
			continue
		}

		// TODO: 重构为一个分发函数
		// 左键点击事件
		if scrcpy.EventType(event.ControlType) == scrcpy.Tapped {
			fmt.Println("Start Tapped event:", event)
			var clickE scrcpy.ClickEvent
			err := json.Unmarshal(message, &clickE)
			if err != nil {
				continue
			}
			GlobalSvc.Client.Tapped(clickE.EventPosition)
		}

		// 鼠标滚动事件
		if scrcpy.EventType(event.ControlType) == scrcpy.Scroll {
			fmt.Println("Start Scroll event:", event)
			var scrollE scrcpy.ScrollEvent
			err := json.Unmarshal(message, &scrollE)
			if err != nil {
				continue
			}

			GlobalSvc.Client.Scroll(scrollE.ScrollPosition)
		}

		// 长按事件
		if scrcpy.EventType(event.ControlType) == scrcpy.Down {
			fmt.Println("Start Press event:", event)
			var downEvent scrcpy.DownEvent
			err := json.Unmarshal(message, &downEvent)
			if err != nil {
				continue
			}
			GlobalSvc.Client.Down(downEvent.EventPosition)
		}

		// 拖拽事件
		if scrcpy.EventType(event.ControlType) == scrcpy.Move {
			fmt.Println("Start Move event:", event)
			var moveE scrcpy.MoveEvent
			err := json.Unmarshal(message, &moveE)
			if err != nil {
				continue
			}

			GlobalSvc.Client.Move(moveE.MovePosition)
		}

		// 拖拽事件
		if scrcpy.EventType(event.ControlType) == scrcpy.Up {
			fmt.Println("Start Up event:", event)
			var upE scrcpy.UpEvent
			err := json.Unmarshal(message, &upE)
			if err != nil {
				continue
			}

			GlobalSvc.Client.Up(upE.EventPosition)
		}
	}
}
