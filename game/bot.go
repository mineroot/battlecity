package game

import (
	"github.com/faiface/pixel"
	"time"
)

type Bot struct {
	model            *Animation
	pos              pixel.Vec
	speed            float64
	direction        Direction
	currentBullet    *Bullet
	lastShootingTime time.Time
	shootingInterval time.Duration
}

func NewBot(spritesheet *pixel.Picture) *Bot {
	p := new(Bot)
	p.pos = pixel.V(10*BlockSize*Scale, 3*BlockSize*Scale)
	p.speed = 44 * Scale
	p.direction = South
	p.shootingInterval = time.Millisecond * 200
	frames := []*pixel.Sprite{
		pixel.NewSprite(*spritesheet, pixel.R(0, 240, 16, 255)),
		pixel.NewSprite(*spritesheet, pixel.R(16, 240, 32, 255)),
	}
	p.model = NewAnimation(frames, 0.07)
	return p
}

func (b *Bot) Side() TankSide {
	return bot
}

func (b *Bot) Pos() pixel.Vec {
	return b.pos
}

func (b *Bot) Direction() Direction {
	return b.direction
}
