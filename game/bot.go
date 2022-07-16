package game

import (
	"github.com/google/uuid"
	"math"
	"math/rand"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

type BotType int

const (
	DefaultBot BotType = iota
	RapidMovementBot
	RapidShootingBot
	ArmoredBot
)

type Bot struct {
	Id
	botType          BotType
	model            *Animation
	pos              pixel.Vec
	speed            float64
	direction        Direction
	hp               int
	currentBullet    *Bullet
	bulletSpeed      float64
	lastShootingTime time.Time
	shootingInterval time.Duration
	maxStuckInterval time.Duration
	stuckTime        time.Duration
}

func NewBot(spritesheet pixel.Picture, botType BotType, pos pixel.Vec) *Bot {
	b := new(Bot)
	b.id = uuid.New()
	b.pos = pos
	b.direction = South
	b.shootingInterval = time.Millisecond * 200
	b.maxStuckInterval = time.Millisecond * 300
	b.stuckTime = 0
	b.initBotType(spritesheet, botType)
	return b
}

func (b *Bot) CalculateMovement(_ *pixelgl.Window, dt float64) (pixel.Vec, Direction) {
	const (
		directionChangeProb = 0.5 // 50% per second
		turnProb            = 0.7 // 70% per direction change
	)
	newDirection := b.direction
	if b.stuckTime > b.maxStuckInterval || directionChangeProb*dt > rand.Float64() {
		b.stuckTime = 0
		if turnProb > rand.Float64() {
			var perpendicularDirections []Direction
			if b.direction.IsHorizontal() {
				perpendicularDirections = []Direction{North, South}
			} else {
				perpendicularDirections = []Direction{West, East}
			}
			newDirection = perpendicularDirections[rand.Intn(len(perpendicularDirections))]
		} else {
			for {
				randomDirection := RandomDirection()
				if randomDirection != b.direction {
					newDirection = randomDirection
					break
				}
			}
		}
	}

	speed := b.speed * dt
	newPos := b.pos.Add(newDirection.Velocity(speed))

	if b.direction.IsPerpendicular(newDirection) {
		switch b.direction {
		case North, South:
			newPos.Y = MRound(math.Round, newPos.Y, Scale*BlockSize)
		case East, West:
			newPos.X = MRound(math.Round, newPos.X, Scale*BlockSize)
		}
	}
	return newPos, newDirection
}

func (b *Bot) Move(movementRes *MovementResult, dt float64) {
	b.direction = movementRes.direction
	if movementRes.canMove {
		b.pos = movementRes.newPos
	} else {
		b.stuckTime += time.Duration(dt * float64(time.Second))
		// alignment
		if b.direction.IsHorizontal() {
			b.pos = pixel.V(MRound(math.Round, b.pos.X, Scale*BlockSize), movementRes.newPos.Y)
		} else {
			b.pos = pixel.V(movementRes.newPos.X, MRound(math.Round, b.pos.Y, Scale*BlockSize))
		}
	}
}

func (b *Bot) Shoot(_ *pixelgl.Window, dt float64) *Bullet {
	const shootProb = 1 // 100% per second
	now := time.Now()
	canShoot := now.Sub(b.lastShootingTime) > b.shootingInterval
	noCurrentBullet := b.currentBullet == nil || b.currentBullet.destroyed
	if noCurrentBullet && canShoot && shootProb*dt > rand.Float64() {
		bullet := CreateBullet(b, b.bulletSpeed)
		b.currentBullet = bullet
		b.lastShootingTime = now
		return bullet
	}

	return nil
}

func (b *Bot) Draw(win *pixelgl.Window, dt float64) {
	frame := b.model.CurrentFrame(dt)
	m := pixel.IM.Moved(b.pos)
	if b.direction > East { // reflect
		m = m.Rotated(b.pos, -math.Pi).
			ScaledXY(b.pos, pixel.V(-1, 1)).
			Rotated(b.pos, math.Pi)
	}
	m = m.Scaled(b.pos, Scale).
		Rotated(b.pos, b.direction.Angle())

	frame.Draw(win, m)
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

func (b *Bot) initBotType(spritesheet pixel.Picture, botType BotType) {
	var frames []*pixel.Sprite
	var speed, bulletSpeed float64
	var hp int
	b.botType = botType
	switch b.botType {
	case DefaultBot:
		frames = []*pixel.Sprite{
			pixel.NewSprite(spritesheet, pixel.R(128, 176, 144, 192)),
			pixel.NewSprite(spritesheet, pixel.R(144, 176, 160, 192)),
		}
		speed, bulletSpeed = 30*Scale, 100*Scale
		hp = 1
	case RapidMovementBot:
		frames = []*pixel.Sprite{
			pixel.NewSprite(spritesheet, pixel.R(128, 160, 144, 176)),
			pixel.NewSprite(spritesheet, pixel.R(144, 160, 160, 176)),
		}
		speed, bulletSpeed = 60*Scale, 100*Scale
		hp = 1
	case RapidShootingBot:
		frames = []*pixel.Sprite{
			pixel.NewSprite(spritesheet, pixel.R(128, 144, 144, 160)),
			pixel.NewSprite(spritesheet, pixel.R(144, 144, 160, 160)),
		}
		speed, bulletSpeed = 30*Scale, 175*Scale
		hp = 1
	case ArmoredBot:
		frames = []*pixel.Sprite{
			pixel.NewSprite(spritesheet, pixel.R(128, 128, 144, 144)),
			pixel.NewSprite(spritesheet, pixel.R(144, 128, 160, 144)),
		}
		speed, bulletSpeed = 30*Scale, 100*Scale
		hp = 4
	}

	b.model = NewAnimation(frames, 0.07)
	b.speed, b.bulletSpeed = speed, bulletSpeed
	b.hp = hp
}
