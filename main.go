package main

import (
	"context"
	"image"
	"sync"
	"xplayer/router"
	"xplayer/scrcpy"
	"xplayer/service"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/xmsociety/adbutils"
)

func main() {
	ch := make(chan image.Image, 1000)
	errCh := make(chan error)
	errChService := make(chan error)
	svc := service.NewService(NewClient("8cf051cc", ch, errCh), errChService, "config")
	go func() {
		svc.Start()
	}()
	r := router.NewRouter()
	r.Run(":8080")
	// err := endless.ListenAndServe(":8080", r) // TODO: 从配置文件获取
	// if err != nil {
	// 	panic(err)
	// }
}

func NewClient(sn string, VideoTransfer chan image.Image, ErrReceiver chan error) *scrcpy.Client {
	if sn == "" {
		sn = "127.0.0.1:5555"
	}
	snNtid := adbutils.SerialNTransportID{
		Serial: sn,
	}
	return &scrcpy.Client{Device: adb.Device(snNtid), MaxWith: service.MaxWidth, MaxFps: 15, Bitrate: 2000000, VideoSender: VideoTransfer, ErrReceiver: ErrReceiver}
}

var (
	headersMap = map[int]string{
		0: "id",
		1: "check",
		2: "SerialNum",
		3: "NickName",
		4: "RunMode",
		5: "Run",
		6: "Operate",
		7: "Other",
	}
	devicesList      = make([]map[int]interface{}, 0)
	adb              = adbutils.AdbClient{Host: "localhost", Port: 5037, SocketTime: 10}
	textMap          = make(map[string]map[string]string)
	LiveMap          = make(map[string]fyne.Window)
	themeSettingOn   = false
	editMap          = make(map[string]fyne.Window)
	clientMap        = make(map[string]*scrcpy.Client)
	clientCancelMap  = make(map[string]context.CancelFunc)
	serviceMap       = make(map[string]*service.Service)
	checkBoxMap      = make(map[string]*widget.Check)
	serviceButtonMap = make(map[string]*widget.Button)
	allCheck         = &widget.Check{}
	allStartBtn      = &widget.Button{}
	allStopBtn       = &widget.Button{}
	opLock           = sync.Mutex{}
	maxWidthCol2     = 150
	maxWidthCol3     = 150
)
