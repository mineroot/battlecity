package entity

import (
	"battlecity2/core"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"math"
	"time"
)

const PlayerSize = 16.0

type Player struct {
	Id
	model          *core.Animation
	pos            pixel.Vec
	scale          float64
	speed          float64
	direction      Direction
	shootingTicker *time.Ticker
	shootingPeriod time.Duration
	canShoot       bool
}

func NewPlayer(spritesheet pixel.Picture, scale float64) *Player {
	p := new(Player)
	p.scale = scale
	p.pos = pixel.V(11*BlockSize*p.scale, 3*BlockSize*p.scale)
	p.speed = 44 * p.scale
	p.direction = North
	frames := []*pixel.Sprite{
		pixel.NewSprite(spritesheet, pixel.R(0, 240, 16, 255)),
		pixel.NewSprite(spritesheet, pixel.R(16, 240, 32, 255)),
	}
	p.model = core.NewAnimation(frames, 0.07)
	p.shootingPeriod = time.Second * 2
	p.shootingTicker = time.NewTicker(p.shootingPeriod)
	p.canShoot = true
	return p
}

func (p *Player) HandleMovementInput(win *pixelgl.Window, dt float64) (pixel.Vec, Direction) {
	var newDirection Direction
	if win.Pressed(pixelgl.KeyW) {
		newDirection = North
	} else if win.Pressed(pixelgl.KeyD) {
		newDirection = East
	} else if win.Pressed(pixelgl.KeyS) {
		newDirection = South
	} else if win.Pressed(pixelgl.KeyA) {
		newDirection = West
	} else {
		return p.pos, p.direction
	}
	speed := p.speed * dt
	newPos := p.pos.Add(newDirection.Velocity(speed))

	// if direction changed by π/2 (i.g. from West to North, but not from South to North)
	if math.Mod(float64(newDirection+p.direction), 2) != 0 {
		switch p.direction {
		case North, South:
			newPos.Y = mRound(math.Round, newPos.Y, p.scale*BlockSize)
		case East, West:
			newPos.X = mRound(math.Round, newPos.X, p.scale*BlockSize)
		}
	}
	return newPos, newDirection
}

func (p *Player) HandleShootingInput(win *pixelgl.Window) *Bullet {
	select {
	case <-p.shootingTicker.C:
		p.canShoot = true
	default:
	}

	if p.canShoot && win.JustPressed(pixelgl.KeySpace) {
		p.shootingTicker.Reset(p.shootingPeriod)
		p.canShoot = false
		return CreateBullet(p, p.scale)
	}

	return nil
}

func (p *Player) Move(pos pixel.Vec, direction Direction) {
	p.pos = pos
	p.direction = direction
}

func (p *Player) Draw(win *pixelgl.Window, dt float64) {
	frame := p.model.CurrentFrame(dt)

	m := pixel.IM.Moved(p.pos)
	if p.direction > East { // reflect
		m = m.Rotated(p.pos, -math.Pi).
			ScaledXY(p.pos, pixel.V(-1, 1)).
			Rotated(p.pos, math.Pi)
	}
	m = m.Scaled(p.pos, p.scale).
		Rotated(p.pos, p.direction.Angle())

	frame.Draw(win, m)
}

func (p *Player) Side() TankSide {
	return Human
}

func (p *Player) Pos() pixel.Vec {
	return p.pos
}

func (p *Player) Direction() Direction {
	return p.direction
}

func mRound(rounder func(float64) float64, n float64, multiple float64) float64 {
	return multiple * rounder(n/multiple)
}
