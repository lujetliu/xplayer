package service

import (
	"context"
	"sync"

	"fyne.io/fyne"
	"fyne.io/fyne/widget"
	"github.com/xmsociety/adbutils"
)

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
	devicesList    = make([]map[int]interface{}, 0)
	adb            = adbutils.AdbClient{Host: "localhost", Port: 5037, SocketTime: 10}
	textMap        = make(map[string]map[string]string)
	LiveMap        = make(map[string]fyne.Window)
	themeSettingOn = false
	editMap        = make(map[string]fyne.Window)
	// clientMap        = make(map[string]*scrcpy.Client)
	clientCancelMap  = make(map[string]context.CancelFunc)
	serviceMap       = make(map[string]*Service)
	checkBoxMap      = make(map[string]*widget.Check)
	serviceButtonMap = make(map[string]*widget.Button)
	allCheck         = &widget.Check{}
	allStartBtn      = &widget.Button{}
	allStopBtn       = &widget.Button{}
	opLock           = sync.Mutex{}
	maxWidthCol2     = 150
	maxWidthCol3     = 150
)
