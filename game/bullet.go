package game

import (
	"battlecity/game/utils"

	"github.com/faiface/pixel"
)

const (
	BulletW = 3.0
	BulletH = 4.0
)

type Bullet struct {
	origin    Tank
	pos       pixel.Vec
	direction utils.Direction
	speed     float64
	destroyed bool
}

func CreateBullet(origin Tank, speed float64) *Bullet {
	b := new(Bullet)
	b.origin = origin
	b.pos = b.origin.Pos().Add(b.origin.Direction().Velocity(TankSize / 2 * Scale))
	b.direction = b.origin.Direction()
	b.speed = speed
	return b
}

func (b *Bullet) Move(dt float64) {
	speed := b.speed * dt
	b.pos = b.pos.Add(b.direction.Velocity(speed))
}

func (b *Bullet) Destroy() {
	b.destroyed = true
}

func (b *Bullet) IsUpgraded() bool {
	player, ok := b.origin.(*Player)
	return ok && player.level == maxLevel
}
