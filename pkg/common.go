package pkg

import (
	"fmt"
	"time"
)

type DebugLevel int

const (
	DebugLevelMin DebugLevel = iota
	DebugLevelError
	DebugLevelWarn
	DebugLevelInfo
	DebugLevelDebug
	DebugLevelMax
)

func DebugLevelWrap(l int) DebugLevel {
	return DebugLevel(l % 6)
}

var debugOpt = DebugLevelMin

type MetaData struct {
	Size   Size
	Format string
}

type Size struct {
	Width  int
	Height int
}

func (s Size) Center() Point {
	return Point{s.Width >> 1, s.Height >> 1}
}

func (s Size) String() string {
	return fmt.Sprintf("size: (%d, %d)", s.Width, s.Height)
}

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (p Point) String() string {
	return fmt.Sprintf("Point: (%d, %d)", p.X, p.Y)
}

type PointMacro struct {
	Point
	Interval time.Duration
}

type SPoint Point

type UserOperation interface {
}
