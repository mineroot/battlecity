package game

import (
	"github.com/faiface/pixel"
	"github.com/google/uuid"
)

const Scale = 4.0

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
}
