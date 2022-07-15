package game

import (
	"math"
	"math/rand"

	"github.com/faiface/pixel"
)

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

func (d Direction) IsPerpendicular(d2 Direction) bool {
	return math.Mod(float64(d2+d), 2) != 0
}

func RandomDirection() Direction {
	return Direction(rand.Intn(4))
}
