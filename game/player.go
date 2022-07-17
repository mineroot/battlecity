package game

import (
	"math"
	"time"

	"battlecity/game/utils"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/google/uuid"
)

type Player struct {
	Id
	model               *utils.Animation
	immunityModel       *utils.Animation
	pos                 pixel.Vec
	speed               float64
	direction           utils.Direction
	immune              bool
	immunityDuration    time.Duration
	maxImmunityDuration time.Duration
	bulletSpeed         float64
	currentBullet       *Bullet
	lastShootingTime    time.Time
	shootingInterval    time.Duration
}

func NewPlayer(spritesheet pixel.Picture) *Player {
	p := new(Player)
	p.id = uuid.New()
	p.Respawn()
	p.shootingInterval = time.Millisecond * 200
	p.model = utils.NewAnimation([]utils.AnimationFrame{
		{
			Frame:    pixel.NewSprite(spritesheet, pixel.R(0, 240, 16, 256)),
			Duration: time.Microsecond * 66666,
		},
		{
			Frame:    pixel.NewSprite(spritesheet, pixel.R(16, 240, 32, 256)),
			Duration: time.Microsecond * 66666,
		},
	})
	p.immunityModel = utils.NewAnimation([]utils.AnimationFrame{
		{
			Frame:    pixel.NewSprite(spritesheet, pixel.R(256, 96, 272, 112)),
			Duration: time.Millisecond * 40,
		},
		{
			Frame:    pixel.NewSprite(spritesheet, pixel.R(272, 96, 288, 112)),
			Duration: time.Millisecond * 40,
		},
	})
	return p
}

func (p *Player) Update(dt float64) {
	if p.immunityDuration >= p.maxImmunityDuration {
		p.immune = false
		p.immunityModel.Reset()
		p.immunityDuration = 0
	}
	if p.immune {
		p.immunityDuration += time.Duration(dt * float64(time.Second))
	}
}

func (p *Player) Respawn() {
	p.pos = pixel.V(11*BlockSize*Scale, 3*BlockSize*Scale)
	p.speed = 44 * Scale
	p.direction = utils.North
	p.MakeImmune(time.Second * 3)
	p.bulletSpeed = 100 * Scale
	p.currentBullet = nil
}

func (p *Player) CalculateMovement(win *pixelgl.Window, dt float64) (pixel.Vec, utils.Direction) {
	var newDirection utils.Direction
	if win.Pressed(pixelgl.KeyW) {
		newDirection = utils.North
	} else if win.Pressed(pixelgl.KeyD) {
		newDirection = utils.East
	} else if win.Pressed(pixelgl.KeyS) {
		newDirection = utils.South
	} else if win.Pressed(pixelgl.KeyA) {
		newDirection = utils.West
	} else {
		return p.pos, p.direction
	}
	speed := p.speed * dt
	newPos := p.pos.Add(newDirection.Velocity(speed))

	if p.direction.IsPerpendicular(newDirection) {
		switch p.direction {
		case utils.North, utils.South:
			newPos.Y = MRound(math.Round, newPos.Y, Scale*BlockSize)
		case utils.East, utils.West:
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
		bullet := CreateBullet(p, p.bulletSpeed)
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

func (p *Player) MakeImmune(maxDuration time.Duration) {
	p.immune = true
	p.maxImmunityDuration = maxDuration
}

func (p *Player) Draw(win *pixelgl.Window, dt float64, isPaused bool) {
	immunityDt := dt
	if isPaused {
		dt = 0
	}
	movementPressed := win.Pressed(pixelgl.KeyW) || win.Pressed(pixelgl.KeyD) ||
		win.Pressed(pixelgl.KeyS) || win.Pressed(pixelgl.KeyA)
	if !movementPressed {
		dt = 0
	}
	frame := p.model.CurrentFrame(dt)

	m := pixel.IM.Moved(p.pos)
	if p.direction > utils.East { // reflect
		m = m.Rotated(p.pos, -math.Pi).
			ScaledXY(p.pos, pixel.V(-1, 1)).
			Rotated(p.pos, math.Pi)
	}
	m = m.Scaled(p.pos, Scale).
		Rotated(p.pos, p.direction.Angle())

	frame.Draw(win, m)
	if p.immune {
		immunityFrame := p.immunityModel.CurrentFrame(immunityDt)
		immunityFrame.Draw(win, pixel.IM.Moved(p.pos).Scaled(p.pos, Scale))
	}
}

func (p *Player) Side() TankSide {
	return human
}

func (p *Player) Pos() pixel.Vec {
	return p.pos
}

func (p *Player) Direction() utils.Direction {
	return p.direction
}
