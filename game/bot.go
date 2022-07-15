package game

import (
	"math"
	"math/rand"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

type Bot struct {
	Id
	model            *Animation
	pos              pixel.Vec
	speed            float64
	direction        Direction
	currentBullet    *Bullet
	lastShootingTime time.Time
	shootingInterval time.Duration
	maxStuckInterval time.Duration
	stuckTime        time.Duration
}

func NewBot(spritesheet pixel.Picture, pos pixel.Vec) *Bot {
	b := new(Bot)
	b.pos = pos
	b.speed = 44 * Scale
	b.direction = South
	b.shootingInterval = time.Millisecond * 200
	b.maxStuckInterval = time.Millisecond * 200
	b.stuckTime = 0
	frames := []*pixel.Sprite{
		pixel.NewSprite(spritesheet, pixel.R(0, 240, 16, 255)),
		pixel.NewSprite(spritesheet, pixel.R(16, 240, 32, 255)),
	}
	b.model = NewAnimation(frames, 0.07)
	return b
}

func (b *Bot) HandleMovement(_ *pixelgl.Window, dt float64) (pixel.Vec, Direction) {
	const (
		directionChangeProb = 0.6 // 60% per second
		turnProb            = 0.7 // 70% per direction change
	)
	newDirection := b.direction
	if b.stuckTime > b.maxStuckInterval || directionChangeProb*dt > rand.Float64() {
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

func (b *Bot) Move(movementRes *MovementResult, dt float64) { // TODO refactor (same as Player.Move)
	b.direction = movementRes.direction
	if movementRes.canMove {
		b.pos = movementRes.newPos
		b.stuckTime = 0
	} else {
		b.stuckTime += time.Duration(dt * float64(time.Second))
		// alignment
		if b.direction.IsHorizontal() {
			b.pos.Y = movementRes.newPos.Y
		} else {
			b.pos.X = movementRes.newPos.X
		}
	}
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
