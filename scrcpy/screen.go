package scrcpy

import (
	"fmt"
	"log"
)

type EventType uint

const (
	Tapped       EventType = 100 + iota // 左键点击
	DoubleTapped                        // 左键双击
	Scroll                              // 滚动滚轮
	Down                                // 长按
	Move                                // 移动
	Up                                  // 拖拽
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

type UpEvent struct {
	Event
	EventPosition
}

type DownEvent struct {
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
	Event
	EventPosition
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
func (client *Client) Down(e EventPosition) {
	fmt.Println("Down", e)
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
}

// 拖拽
func (client *Client) Move(e MovePosition) {
	fmt.Println("Move", e)
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	if client.Control.ControlConn == nil {
		log.Println("o.Client.Control.ControlConn is nil")
		return
	}
	client.Control.Touch(int(e.X), int(e.Y), ActionMove)
}

func (client *Client) Up(e EventPosition) {
	fmt.Println("Down", e)
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	if client.Control.ControlConn == nil {
		log.Println("o.Client.Control.ControlConn is nil")
		return
	}
	client.Control.Touch(int(e.X), int(e.Y), ActionUp)
}
