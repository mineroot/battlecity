package game

import (
	"github.com/faiface/pixel"
	"github.com/google/uuid"
	"math"
)

const Scale = 4.0

type TankSide int

const (
	Human TankSide = iota
	Bot
)

type Id struct {
	id uuid.UUID
}

func (i *Id) Id() uuid.UUID {
	if i.id.ID() == 0 {
		i.id = uuid.New()
	}
	return i.id
}

type Tank interface {
	Side() TankSide
	Pos() pixel.Vec
	Direction() Direction
}

type Direction int

const (
	North Direction = iota
	East
	South
	West
)

func (d Direction) Velocity(speed float64) pixel.Vec {
	var vec pixel.Vec
	switch d {
	case North:
		vec = pixel.V(0, 1)
	case East:
		vec = pixel.V(1, 0)
	case South:
		vec = pixel.V(0, -1)
	case West:
		vec = pixel.V(-1, 0)
	}
	return vec.Scaled(speed)
}

func (d Direction) Angle() float64 {
	switch d {
	case North:
		return 0
	case East:
		return 3 * math.Pi / 2
	case South:
		return math.Pi
	case West:
		return math.Pi / 2
	default:
		panic("direction: unreachable statement")
	}
}

func (d Direction) IsHorizontal() bool {
	return math.Mod(float64(d), 2) == 1
}

func (d Direction) IsVertical() bool {
	return !d.IsHorizontal()
}
