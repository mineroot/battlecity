package game

import (
	"battlecity/game/sfx"
	"battlecity/game/utils"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/google/uuid"
	"math"
	"time"
)

const maxLevel = 3

type Player struct {
	Id
	spritesheet         pixel.Picture
	model               *utils.Animation
	immunityModel       *utils.Animation
	creationModel       *utils.Animation
	pos                 pixel.Vec
	speed               float64
	direction           utils.Direction
	immune              bool
	onCreation          bool
	immunityDuration    time.Duration
	maxImmunityDuration time.Duration
	bulletSpeed         float64
	currentBullet1      *Bullet
	currentBullet2      *Bullet
	lastShootingTime    time.Time
	shootingInterval    time.Duration
	level               int
	lives               int
}

func NewPlayer(spritesheet pixel.Picture) *Player {
	p := new(Player)
	p.id = uuid.New()
	p.spritesheet = spritesheet
	p.lives = 2
	p.shootingInterval = time.Millisecond * 200
	p.immunityModel = utils.NewAnimation([]utils.AnimationFrame{
		{
			Frame:    pixel.NewSprite(spritesheet, pixel.R(256, 96, 272, 112)),
			Duration: time.Millisecond * 40,
		},
		{
			Frame:    pixel.NewSprite(spritesheet, pixel.R(272, 96, 288, 112)),
			Duration: time.Millisecond * 40,
		},
	}, -1)
	creationAnimationSprites := []*pixel.Sprite{
		pixel.NewSprite(spritesheet, pixel.R(256, 144, 272, 160)),
		pixel.NewSprite(spritesheet, pixel.R(272, 144, 288, 160)),
		pixel.NewSprite(spritesheet, pixel.R(288, 144, 304, 160)),
		pixel.NewSprite(spritesheet, pixel.R(304, 144, 320, 160)),
	}
	creationFramesSeq := []int{3, 2, 1, 0, 1, 2, 3, 2, 1, 0, 1, 2, 3}
	creationFrames := make([]utils.AnimationFrame, len(creationFramesSeq))
	creationAnimationDuration := time.Millisecond * 40
	for i, creationFrameI := range creationFramesSeq {
		creationFrames[i] = utils.AnimationFrame{
			Frame:    creationAnimationSprites[creationFrameI],
			Duration: creationAnimationDuration,
		}
	}
	p.creationModel = utils.NewAnimation(creationFrames, 1)
	return p
}

func (p *Player) Update(dt float64) {
	if p.onCreation {
		return
	}
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
	p.MakeImmune(time.Second * 3)
	p.onCreation = true
	p.pos = pixel.V(11*BlockSize*Scale, 3*BlockSize*Scale)
	p.direction = utils.North
	p.currentBullet1 = nil
	p.currentBullet2 = nil
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
		sfx.PlayTankIdle()
		return p.pos, p.direction
	}
	sfx.PlayTankMoving()
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
	if p.onCreation {
		return nil
	}
	now := time.Now()
	if p.currentBullet1 != nil && p.currentBullet1.destroyed {
		p.currentBullet1 = nil
	}
	if p.currentBullet2 != nil && p.currentBullet2.destroyed {
		p.currentBullet2 = nil
	}

	var shootingInterval time.Duration
	var canShoot bool
	if p.level < 2 {
		shootingInterval = p.shootingInterval
		canShoot = p.currentBullet1 == nil
	} else {
		canShoot = p.currentBullet2 == nil
		if canShoot {
			shootingInterval = p.shootingInterval / 2
			if p.currentBullet1 != nil && p.currentBullet2 == nil {
				shootingInterval = p.shootingInterval / 8
			}
		}
	}
	canShoot = canShoot && now.Sub(p.lastShootingTime) >= shootingInterval
	if canShoot && win.JustPressed(pixelgl.KeySpace) {
		sfx.PlayShoot()
		bullet := CreateBullet(p, p.bulletSpeed)
		if p.currentBullet1 == nil {
			p.currentBullet1 = bullet
		} else {
			p.currentBullet2 = bullet
		}
		p.lastShootingTime = now
		return bullet
	}

	return nil
}

func (p *Player) Move(movementRes *MovementResult, _ float64) {
	if p.onCreation {
		return
	}
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

func (p *Player) Upgrade() {
	if p.level != maxLevel {
		p.changeLevel(p.level + 1)
	}
}

func (p *Player) ResetLevel() {
	p.changeLevel(0)
}

func (p *Player) Draw(win *pixelgl.Window, dt float64, isPaused bool) {
	immunityDt := dt
	if isPaused {
		dt = 0
	}
	if p.onCreation {
		frame := p.creationModel.CurrentFrame(dt)
		if frame != nil {
			m := pixel.IM.Moved(p.pos).Scaled(p.pos, Scale)
			frame.Draw(win, m)
			return
		} else {
			p.onCreation = false
			p.creationModel.Reset()
		}
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

func (p *Player) OnCreation() bool {
	return p.onCreation
}

func (p *Player) changeLevel(level int) {
	if level >= 4 {
		panic("player: level out of bounds [0, 4)")
	}
	p.level = level

	switch p.level {
	case 0:
		p.bulletSpeed = 100 * Scale
		p.speed = 44 * Scale
	default:
		p.bulletSpeed = 200 * Scale
		p.speed = 50 * Scale
	}

	minYStart, maxYStart := 240.0, 256.0
	minY, maxY := minYStart-float64(p.level)*TankSize, maxYStart-float64(p.level)*TankSize
	p.model = utils.NewAnimation([]utils.AnimationFrame{
		{
			Frame:    pixel.NewSprite(p.spritesheet, pixel.R(0, minY, 16, maxY)),
			Duration: time.Microsecond * 66666,
		},
		{
			Frame:    pixel.NewSprite(p.spritesheet, pixel.R(16, minY, 32, maxY)),
			Duration: time.Microsecond * 66666,
		},
	}, -1)
}
