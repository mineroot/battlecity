package entity

import (
	"github.com/faiface/pixel"
)

const (
	BorderBlock = "|"
	BrickBlock  = "b"
	SteelBlock  = "s"
	WaterBlock  = "w"
	SpaceBlock  = " "
	BlockSize   = 8.0
)

type Block struct {
	kind          string
	destroyable   bool // can bullet destroy it
	passable      bool // can tank pass through it
	shootable     bool // can bullet pass through it
	pos           pixel.Vec
	quadrants     [2][2]bool
	quadrantRects [2][2]pixel.Rect
}

func Border(pos pixel.Vec) *Block {
	block := new(Block)
	block.pos = pos
	block.kind = BorderBlock
	block.destroyable = false
	block.passable = false
	block.shootable = false
	return block
}

func Brick(pos pixel.Vec) *Block {
	block := new(Block)
	block.pos = pos
	block.kind = BrickBlock
	block.destroyable = true
	block.passable = false
	block.shootable = false
	block.quadrants = [2][2]bool{{true, true}, {false, true}}
	return block
}

func Steel(pos pixel.Vec) *Block {
	block := new(Block)
	block.pos = pos
	block.kind = SteelBlock
	block.destroyable = false
	block.passable = false
	block.shootable = false
	return block
}

func Water(pos pixel.Vec) *Block {
	block := new(Block)
	block.pos = pos
	block.kind = WaterBlock
	block.destroyable = false
	block.passable = false
	block.shootable = true
	return block
}

func Space(pos pixel.Vec) *Block {
	block := new(Block)
	block.pos = pos
	block.kind = SpaceBlock
	block.destroyable = false
	block.passable = true
	block.shootable = true
	return block
}

func (b *Block) SetQuadrantRects(quadrantRects [2][2]pixel.Rect) {
	b.quadrantRects = quadrantRects
}

func (b *Block) QuadrantRect(i, j int) *pixel.Rect {
	if !b.quadrants[i][j] {
		return &b.quadrantRects[i][j]
	}
	return nil
}

func (b *Block) Pos() pixel.Vec {
	return b.pos
}
func (b *Block) Kind() string {
	return b.kind
}

func (b *Block) Passable() bool {
	return b.passable
}
