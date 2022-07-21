package game

import (
	"battlecity/game/utils"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/google/uuid"
	"math"
	"math/rand"
	"time"
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
	model            *utils.Animation
	bonusModel       *utils.Animation
	bonusModelPaused *utils.Animation
	pos              pixel.Vec
	isBonus          bool
	speed            float64
	direction        utils.Direction
	hp               int
	currentBullet    *Bullet
	bulletSpeed      float64
	lastShootingTime time.Time
	shootingInterval time.Duration
	maxStuckInterval time.Duration
	stuckTime        time.Duration
}

func NewBot(spritesheet pixel.Picture, botType BotType, pos pixel.Vec, isBonus bool) *Bot {
	b := new(Bot)
	b.id = uuid.New()
	b.pos = pos
	b.isBonus = isBonus
	b.direction = utils.South
	b.shootingInterval = time.Millisecond * 200
	b.maxStuckInterval = time.Millisecond * 300
	b.stuckTime = 0
	b.initBotType(spritesheet, botType)
	return b
}

func (b *Bot) CalculateMovement(_ *pixelgl.Window, dt float64) (pixel.Vec, utils.Direction) {
	const (
		directionChangeProb = 0.5 // 50% per second
		turnProb            = 0.7 // 70% per direction change
	)
	newDirection := b.direction
	if b.stuckTime > b.maxStuckInterval || directionChangeProb*dt > rand.Float64() {
		b.stuckTime = 0
		if turnProb > rand.Float64() {
			var perpendicularDirections []utils.Direction
			if b.direction.IsHorizontal() {
				perpendicularDirections = []utils.Direction{utils.North, utils.South}
			} else {
				perpendicularDirections = []utils.Direction{utils.West, utils.East}
			}
			newDirection = perpendicularDirections[rand.Intn(len(perpendicularDirections))]
		} else {
			for {
				randomDirection := utils.RandomDirection()
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
		case utils.North, utils.South:
			newPos.Y = MRound(math.Round, newPos.Y, Scale*BlockSize)
		case utils.East, utils.West:
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

func (b *Bot) Draw(win *pixelgl.Window, dt float64, isPaused bool) {
	var frame *pixel.Sprite
	if b.isBonus {
		if isPaused {
			frame = b.bonusModelPaused.CurrentFrame(dt)
		} else {
			frame = b.bonusModel.CurrentFrame(dt)
		}
	} else {
		if isPaused {
			dt = 0
		}
		frame = b.model.CurrentFrame(dt)
	}
	m := pixel.IM.Moved(b.pos)
	if b.direction > utils.East { // reflect
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

func (b *Bot) Direction() utils.Direction {
	return b.direction
}

func (b *Bot) initBotType(spritesheet pixel.Picture, botType BotType) {
	var frames []utils.AnimationFrame
	var speed, bulletSpeed float64
	var hp int
	b.botType = botType
	duration := time.Microsecond * 66666
	switch b.botType {
	case DefaultBot:
		defaultBotFrame := pixel.NewSprite(spritesheet, pixel.R(128, 176, 144, 192))
		defaultBotFrame2 := pixel.NewSprite(spritesheet, pixel.R(144, 176, 160, 192))
		defaultBotBonusFrame := pixel.NewSprite(spritesheet, pixel.R(128, 48, 144, 64))
		defaultBotBonusFrame2 := pixel.NewSprite(spritesheet, pixel.R(144, 48, 160, 64))
		frames = []utils.AnimationFrame{
			{Frame: defaultBotFrame, Duration: duration},
			{Frame: defaultBotFrame2, Duration: duration},
			{Frame: defaultBotFrame, Duration: duration},
			{Frame: defaultBotFrame2, Duration: duration},
			{Frame: defaultBotBonusFrame, Duration: duration},
			{Frame: defaultBotBonusFrame2, Duration: duration},
			{Frame: defaultBotBonusFrame, Duration: duration},
			{Frame: defaultBotBonusFrame2, Duration: duration},
		}
		speed, bulletSpeed = 30*Scale, 100*Scale
		hp = 1
	case RapidMovementBot:
		rapidMovementBotFrame := pixel.NewSprite(spritesheet, pixel.R(128, 160, 144, 176))
		rapidMovementBotFrame2 := pixel.NewSprite(spritesheet, pixel.R(144, 160, 160, 176))
		rapidMovementBotBonusFrame := pixel.NewSprite(spritesheet, pixel.R(128, 32, 144, 48))
		rapidMovementBotBonusFrame2 := pixel.NewSprite(spritesheet, pixel.R(144, 32, 160, 48))
		frames = []utils.AnimationFrame{
			{Frame: rapidMovementBotFrame, Duration: duration},
			{Frame: rapidMovementBotFrame2, Duration: duration},
			{Frame: rapidMovementBotFrame, Duration: duration},
			{Frame: rapidMovementBotFrame2, Duration: duration},
			{Frame: rapidMovementBotBonusFrame, Duration: duration},
			{Frame: rapidMovementBotBonusFrame2, Duration: duration},
			{Frame: rapidMovementBotBonusFrame, Duration: duration},
			{Frame: rapidMovementBotBonusFrame2, Duration: duration},
		}
		speed, bulletSpeed = 60*Scale, 100*Scale
		hp = 1
	case RapidShootingBot:
		rapidShootingBotFrame := pixel.NewSprite(spritesheet, pixel.R(128, 144, 144, 160))
		rapidShootingBotFrame2 := pixel.NewSprite(spritesheet, pixel.R(144, 144, 160, 160))
		rapidShootingBotBonusFrame := pixel.NewSprite(spritesheet, pixel.R(128, 16, 144, 32))
		rapidShootingBotBonusFrame2 := pixel.NewSprite(spritesheet, pixel.R(144, 16, 160, 32))
		frames = []utils.AnimationFrame{
			{Frame: rapidShootingBotFrame, Duration: duration},
			{Frame: rapidShootingBotFrame2, Duration: duration},
			{Frame: rapidShootingBotFrame, Duration: duration},
			{Frame: rapidShootingBotFrame2, Duration: duration},
			{Frame: rapidShootingBotBonusFrame, Duration: duration},
			{Frame: rapidShootingBotBonusFrame2, Duration: duration},
			{Frame: rapidShootingBotBonusFrame, Duration: duration},
			{Frame: rapidShootingBotBonusFrame2, Duration: duration},
		}
		speed, bulletSpeed = 30*Scale, 175*Scale
		hp = 1
	case ArmoredBot:
		armoredBotFrame := pixel.NewSprite(spritesheet, pixel.R(128, 128, 144, 144))
		armoredBotFrame2 := pixel.NewSprite(spritesheet, pixel.R(144, 128, 160, 144))
		armoredBotBonusFrame := pixel.NewSprite(spritesheet, pixel.R(128, 0, 144, 16))
		armoredBotBonusFrame2 := pixel.NewSprite(spritesheet, pixel.R(144, 0, 160, 16))
		frames = []utils.AnimationFrame{
			{Frame: armoredBotFrame, Duration: duration},
			{Frame: armoredBotFrame2, Duration: duration},
			{Frame: armoredBotFrame, Duration: duration},
			{Frame: armoredBotFrame2, Duration: duration},
			{Frame: armoredBotBonusFrame, Duration: duration},
			{Frame: armoredBotBonusFrame2, Duration: duration},
			{Frame: armoredBotBonusFrame, Duration: duration},
			{Frame: armoredBotBonusFrame2, Duration: duration},
		}
		speed, bulletSpeed = 30*Scale, 100*Scale
		hp = 4
	}

	b.model = utils.NewAnimation(frames[:2])
	b.bonusModel = utils.NewAnimation(frames)
	b.bonusModelPaused = utils.NewAnimation([]utils.AnimationFrame{
		frames[0], frames[0], frames[2], frames[2], frames[4], frames[4], frames[6], frames[6],
	})
	b.speed, b.bulletSpeed = speed, bulletSpeed
	b.hp = hp
}
