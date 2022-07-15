package game

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/google/uuid"
)

const Scale = 4.0
const TankSize = 16.0

type TankSide int

const (
	human TankSide = iota
	bot
)

type Id struct {
	id uuid.UUID
}

type Tank interface {
	Side() TankSide
	Pos() pixel.Vec
	Direction() Direction
	CalculateMovement(win *pixelgl.Window, dt float64) (pixel.Vec, Direction)
	Move(movementRes *MovementResult, dt float64)
	Shoot(win *pixelgl.Window, dt float64) *Bullet
}

type MovementResult struct {
	newPos    pixel.Vec
	direction Direction
	canMove   bool
}
