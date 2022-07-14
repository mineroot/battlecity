package entity

import "github.com/faiface/pixel"

type Bullet struct {
	scale     float64
	origin    Tank
	pos       pixel.Vec
	direction Direction
	speed     float64
}

func CreateBullet(origin Tank, scale float64) *Bullet {
	b := new(Bullet)
	b.scale = scale
	b.origin = origin
	b.pos = b.origin.Pos()
	b.direction = b.origin.Direction()
	b.speed = 100 * b.scale
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
