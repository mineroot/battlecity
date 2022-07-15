package game

import (
	"math"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/google/uuid"
)

type Player struct {
	Id
	model            *Animation
	pos              pixel.Vec
	speed            float64
	direction        Direction
	currentBullet    *Bullet
	lastShootingTime time.Time
	shootingInterval time.Duration
}

func NewPlayer(spritesheet pixel.Picture) *Player {
	p := new(Player)
	p.id = uuid.New()
	p.pos = pixel.V(11*BlockSize*Scale, 3*BlockSize*Scale)
	p.speed = 44 * Scale
	p.direction = North
	p.shootingInterval = time.Millisecond * 200
	frames := []*pixel.Sprite{
		pixel.NewSprite(spritesheet, pixel.R(0, 240, 16, 256)),
		pixel.NewSprite(spritesheet, pixel.R(16, 240, 32, 256)),
	}
	p.model = NewAnimation(frames, 0.07)
	return p
}

func (p *Player) CalculateMovement(win *pixelgl.Window, dt float64) (pixel.Vec, Direction) {
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

	if p.direction.IsPerpendicular(newDirection) {
		switch p.direction {
		case North, South:
			newPos.Y = MRound(math.Round, newPos.Y, Scale*BlockSize)
		case East, West:
			newPos.X = MRound(math.Round, newPos.X, Scale*BlockSize)
		}
	}
	return newPos, newDirection
}

func (p *Player) Shoot(win *pixelgl.Window, _ float64) *Bullet {
	now := time.Now()
	canShoot := now.Sub(p.lastShootingTime) > p.shootingInterval
	noCurrentBullet := p.currentBullet == nil || p.currentBullet.destroyed
	if noCurrentBullet && canShoot && win.JustPressed(pixelgl.KeySpace) {
		bullet := CreateBullet(p)
		p.currentBullet = bullet
		p.lastShootingTime = now
		return bullet
	}

	return nil
}

func (p *Player) Move(movementRes *MovementResult, _ float64) {
	p.direction = movementRes.direction
	if movementRes.canMove {
		p.pos = movementRes.newPos
	} else {
		// alignment
		if p.direction.IsHorizontal() {
			p.pos = pixel.V(MRound(math.Round, p.pos.X, Scale*BlockSize), movementRes.newPos.Y)
		} else {
			p.pos = pixel.V(movementRes.newPos.X, MRound(math.Round, p.pos.Y, Scale*BlockSize))
		}
	}
}

func (p *Player) Draw(win *pixelgl.Window, dt float64) {
	movementPressed := win.Pressed(pixelgl.KeyW) || win.Pressed(pixelgl.KeyD) ||
		win.Pressed(pixelgl.KeyS) || win.Pressed(pixelgl.KeyA)
	if !movementPressed {
		dt = 0
	}
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
	return human
}

func (p *Player) Pos() pixel.Vec {
	return p.pos
}

func (p *Player) Direction() Direction {
	return p.direction
}
