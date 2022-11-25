package scrcpy

import (
	"fmt"
	"log"
	"time"
)

type EventType uint

const (
	Tapped       EventType = 100 + iota // 左键点击
	DoubleTapped                        // 左键双击
	Scroll                              // 滚动滚轮
	Press                               // 长按
	Move                                // 拖拽
)

type Event struct {
	IsEvent     uint `json:"is_event"`
	ControlType int  `json:"control_type"`
}

// TODO: 优化
type ClickEvent struct {
	Event
	EventPosition
}

type PressEvent struct {
	Event
	EventPosition
}

type ScrollEvent struct {
	Event
	ScrollPosition
}

type MoveEvent struct {
	Event
	MovePosition
}

type MovePosition struct {
	EventPosition
	EndX       int     `json:"end_x"`
	EndY       int     `json:"end_y"`
	StepLength int     `json:"step_length"`
	Delay      float32 `json:"delay"`
}

// 从前端获取的控制事件时鼠标的位置
type EventPosition struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

type ScrollPosition struct {
	EventPosition
	H float32 `json:"h"`
	V float32 `json:"v"`
}

// 左键点击
func (client *Client) Tapped(e EventPosition) {
	fmt.Println("Tapped", e)
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	if client.Control.ControlConn == nil {
		log.Println("o.Client.Control.ControlConn is nil")
		return
	}
	client.Control.Touch(int(e.X), int(e.Y), ActionDown)
	client.Control.Touch(int(e.X), int(e.Y), ActionUp)
}

// 滚动滚轮
func (client *Client) Scroll(s ScrollPosition) {
	fmt.Println("Scroll", s)
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	if client.Control.ControlConn == nil {
		log.Println("o.Client.Control.ControlConn is nil")
		return
	}
	client.Control.Scroll(int(s.X), int(s.Y), int(s.H), int(s.V))
}

// 长按
func (client *Client) Press(e EventPosition) {
	fmt.Println("Press", e)
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	if client.Control.ControlConn == nil {
		log.Println("o.Client.Control.ControlConn is nil")
		return
	}
	client.Control.Touch(int(e.X), int(e.Y), ActionDown)
	time.Sleep(400 * time.Millisecond)
	client.Control.Touch(int(e.X), int(e.Y), ActionUp)
}

// 拖拽
func (client *Client) Move(e MovePosition) {
	fmt.Println("Press", e)
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	if client.Control.ControlConn == nil {
		log.Println("o.Client.Control.ControlConn is nil")
		return
	}
	client.Control.Touch(int(e.X), int(e.Y), ActionDown)
	time.Sleep(400 * time.Millisecond)
	client.Control.Touch(int(e.EndX), int(e.EndY), ActionUp)
}
