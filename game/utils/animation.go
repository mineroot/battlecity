package utils

import (
	"time"

	"github.com/faiface/pixel"
)

type AnimationFrame struct {
	Frame    *pixel.Sprite
	Duration time.Duration
}

type Animation struct {
	frames            []AnimationFrame
	framesCount       int
	animationDuration time.Duration
	totalDuration     time.Duration
}

func NewAnimation(frames []AnimationFrame) *Animation {
	a := new(Animation)
	a.frames = frames
	a.framesCount = len(a.frames)
	for _, frame := range a.frames {
		a.animationDuration += frame.Duration
	}
	return a
}

func (a *Animation) CurrentFrame(dt float64) *pixel.Sprite {
	dtDuration := time.Duration(dt * float64(time.Second))
	a.totalDuration += dtDuration

	circleNum := a.totalDuration / a.animationDuration
	currentDuration := a.totalDuration - circleNum*a.animationDuration

	framesDuration := time.Duration(0)
	for _, frame := range a.frames {
		framesDuration += frame.Duration
		if framesDuration > currentDuration {
			return frame.Frame
		}
	}
	panic("animation: unreachable statement")
}

func (a *Animation) Reset() {
	a.totalDuration = 0
}
