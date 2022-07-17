package game

import (
	"battlecity/game/utils"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/google/uuid"
)

const Scale = 4.0
const TankSize = 16

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
	Direction() utils.Direction
	CalculateMovement(win *pixelgl.Window, dt float64) (pixel.Vec, utils.Direction)
	Move(movementRes *MovementResult, dt float64)
	Shoot(win *pixelgl.Window, dt float64) *Bullet
}

type MovementResult struct {
	newPos    pixel.Vec
	direction utils.Direction
	canMove   bool
}
