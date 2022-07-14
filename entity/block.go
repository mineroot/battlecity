package entity

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
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
	kind           string
	destroyable    bool // can bullet destroy it
	passable       bool // can tank pass through it
	shootable      bool // can bullet pass through it
	pos            pixel.Vec
	quadrants      [2][2]bool
	quadrantIMDraw [2][2]*imdraw.IMDraw
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

func (b *Block) InitQuadrants(rects [2][2]pixel.Rect) {
	if !b.destroyable {
		return
	}
	for i, rectsI := range rects {
		for j, rect := range rectsI {
			imd := imdraw.New(nil)
			imd.Color = pixel.RGB(0, 0, 0)
			imd.Push(rect.Min, rect.Max)
			imd.Rectangle(0)
			b.quadrantIMDraw[i][j] = imd
		}
	}
}

func (b *Block) QuadrantIMDraw(i, j int) *imdraw.IMDraw {
	if !b.quadrants[i][j] {
		return b.quadrantIMDraw[i][j]
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
