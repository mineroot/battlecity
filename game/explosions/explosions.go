package explosions

import (
	"battlecity/game/utils"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"time"
)

type ExplosionType int

const (
	BulletExplosion ExplosionType = iota
	TankExplosion
)

var (
	scale           float64
	explosionFrames [5]*pixel.Sprite
	initialized     bool
)

func InnitExplosionFrames(spritesheet pixel.Picture, scl float64) {
	if initialized {
		panic("explosions: already initialized")
	}
	scale = scl
	explosionFrames = [5]*pixel.Sprite{
		// 16x16
		pixel.NewSprite(spritesheet, pixel.R(256, 112, 272, 128)),
		pixel.NewSprite(spritesheet, pixel.R(272, 112, 288, 128)),
		pixel.NewSprite(spritesheet, pixel.R(288, 112, 304, 128)),
		// 32x32
		pixel.NewSprite(spritesheet, pixel.R(304, 96, 336, 128)),
		pixel.NewSprite(spritesheet, pixel.R(336, 96, 368, 128)),
	}
	initialized = true
}

type Explosion struct {
	model         *utils.Animation
	isEnded       bool
	pos           pixel.Vec
	explosionType ExplosionType
}

func NewExplosion(explosionType ExplosionType, pos pixel.Vec) *Explosion {
	e := new(Explosion)
	e.explosionType = explosionType
	e.pos = pos
	if e.explosionType == BulletExplosion {
		duration := time.Millisecond * 25
		e.model = utils.NewAnimation([]utils.AnimationFrame{
			{
				Frame:    explosionFrames[0],
				Duration: duration,
			},
			{
				Frame:    explosionFrames[1],
				Duration: duration,
			},
			{
				Frame:    explosionFrames[2],
				Duration: duration,
			},
			{
				Frame:    explosionFrames[1],
				Duration: duration,
			},
			{
				Frame:    explosionFrames[0],
				Duration: duration,
			},
		}, 1)
	} else {
		duration := time.Millisecond * 40
		e.model = utils.NewAnimation([]utils.AnimationFrame{
			{
				Frame:    explosionFrames[0],
				Duration: duration,
			},
			{
				Frame:    explosionFrames[1],
				Duration: duration,
			},
			{
				Frame:    explosionFrames[2],
				Duration: duration,
			},
			{
				Frame:    explosionFrames[3],
				Duration: duration,
			},
			{
				Frame:    explosionFrames[4],
				Duration: duration,
			},
			{
				Frame:    explosionFrames[3],
				Duration: duration,
			},
			{
				Frame:    explosionFrames[2],
				Duration: duration,
			},
			{
				Frame:    explosionFrames[1],
				Duration: duration,
			},
			{
				Frame:    explosionFrames[0],
				Duration: duration,
			},
		}, 1)
	}

	return e
}

func (e *Explosion) IsEnded() bool {
	return e.isEnded
}

func (e *Explosion) Draw(win *pixelgl.Window, dt float64, isPaused bool) {
	if isPaused {
		dt = 0
	}
	frame := e.model.CurrentFrame(dt)
	if frame == nil {
		e.isEnded = true
		return
	}
	frame.Draw(win, pixel.IM.Moved(e.pos).Scaled(e.pos, scale))
}
