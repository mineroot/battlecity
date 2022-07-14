package entity

import "github.com/faiface/pixel"

const (
	BulletW = 3.0
	BulletH = 4.0
)

type Bullet struct {
	origin    Tank
	pos       pixel.Vec
	direction Direction
	speed     float64
}

func CreateBullet(origin Tank) *Bullet {
	b := new(Bullet)
	b.origin = origin
	b.pos = b.origin.Pos()
	b.direction = b.origin.Direction()
	b.speed = 100 * Scale
	return b
}

func (b *Bullet) Move(dt float64) {
	speed := b.speed * dt
	b.pos = b.pos.Add(b.direction.Velocity(speed))
}

func (b *Bullet) Pos() pixel.Vec {
	return b.pos
}

func (b *Bullet) Direction() Direction {
	return b.direction
}
