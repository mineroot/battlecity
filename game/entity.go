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
	HandleMovement(win *pixelgl.Window, dt float64) (pixel.Vec, Direction)
	Move(movementRes *MovementResult, dt float64)
}

type MovementResult struct {
	newPos    pixel.Vec
	direction Direction
	canMove   bool
}
