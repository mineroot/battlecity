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
	model            *core.Animation
	pos              pixel.Vec
	speed            float64
	direction        Direction
	currentBullet    *Bullet
	lastShootingTime time.Time
	shootingInterval time.Duration
}

func NewPlayer(spritesheet pixel.Picture) *Player {
	p := new(Player)
	p.pos = pixel.V(11*BlockSize*Scale, 3*BlockSize*Scale)
	p.speed = 44 * Scale
	p.direction = North
	p.shootingInterval = time.Millisecond * 200
	frames := []*pixel.Sprite{
		pixel.NewSprite(spritesheet, pixel.R(0, 240, 16, 255)),
		pixel.NewSprite(spritesheet, pixel.R(16, 240, 32, 255)),
	}
	p.model = core.NewAnimation(frames, 0.07)
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
			newPos.Y = mRound(math.Round, newPos.Y, Scale*BlockSize)
		case East, West:
			newPos.X = mRound(math.Round, newPos.X, Scale*BlockSize)
		}
	}
	return newPos, newDirection
}

func (p *Player) HandleShootingInput(win *pixelgl.Window) *Bullet {
	now := time.Now()
	canShoot := now.Sub(p.lastShootingTime) > p.shootingInterval
	noCurrentBullet := p.currentBullet == nil || p.currentBullet.IsDestroyed()
	if noCurrentBullet && canShoot && win.JustPressed(pixelgl.KeySpace) {
		bullet := CreateBullet(p)
		p.currentBullet = bullet
		p.lastShootingTime = now
		return bullet
	}

	return nil
}

func (p *Player) Move(canMove bool, pos pixel.Vec, direction Direction) {
	p.direction = direction
	if canMove {
		p.pos = pos
	} else {
		// alignment
		if p.direction.IsHorizontal() {
			p.pos.Y = pos.Y
		} else {
			p.pos.X = pos.X
		}
	}
}

func (p *Player) Draw(win *pixelgl.Window, dt float64) {
	frame := p.model.CurrentFrame(dt)

	m := pixel.IM.Moved(p.pos)
	if p.direction > East { // reflect
		m = m.Rotated(p.pos, -math.Pi).
			ScaledXY(p.pos, pixel.V(-1, 1)).
			Rotated(p.pos, math.Pi)
	}
	m = m.Scaled(p.pos, Scale).
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
