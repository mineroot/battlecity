package game

import (
	"math"

	"github.com/faiface/pixel"
)

type Animation struct {
	frames      []*pixel.Sprite
	framesCount int
	frameTime   float64
	totalTime   float64
}

func NewAnimation(frames []*pixel.Sprite, frameTime float64) *Animation {
	a := new(Animation)
	a.frames = frames
	a.framesCount = len(a.frames)
	a.frameTime = frameTime
	return a
}

func (a *Animation) CurrentFrame(dt float64) *pixel.Sprite {
	a.totalTime += dt
	currentFrame := int(math.Mod(math.Floor(a.totalTime/a.frameTime), float64(a.framesCount)))
	return a.frames[currentFrame]
}
