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

var (
	fullQuadrants  = [2][2]bool{{true, true}, {true, true}}
	emptyQuadrants = [2][2]bool{{false, false}, {false, false}}
)

type Block struct {
	kind           string
	row            int
	column         int
	destroyable    bool // can bullet destroy it
	passable       bool // can tank pass through it
	shootable      bool // can bullet pass through it
	pos            pixel.Vec
	quadrants      [2][2]bool
	quadrantIMDraw [2][2]*imdraw.IMDraw
}

func Border(pos pixel.Vec, row, column int) *Block {
	block := new(Block)
	block.row, block.column = row, column
	block.pos = pos
	block.kind = BorderBlock
	block.destroyable = false
	block.passable = false
	block.shootable = false
	return block
}

func Brick(pos pixel.Vec, row, column int) *Block {
	block := new(Block)
	block.row, block.column = row, column
	block.pos = pos
	block.kind = BrickBlock
	block.destroyable = true
	block.passable = false
	block.shootable = false
	block.quadrants = fullQuadrants
	return block
}

func Steel(pos pixel.Vec, row, column int) *Block {
	block := new(Block)
	block.row, block.column = row, column
	block.pos = pos
	block.kind = SteelBlock
	block.destroyable = false
	block.passable = false
	block.shootable = false
	return block
}

func Water(pos pixel.Vec, row, column int) *Block {
	block := new(Block)
	block.row, block.column = row, column
	block.pos = pos
	block.kind = WaterBlock
	block.destroyable = false
	block.passable = false
	block.shootable = true
	return block
}

func Space(pos pixel.Vec, row, column int) *Block {
	block := new(Block)
	block.row, block.column = row, column
	block.pos = pos
	block.kind = SpaceBlock
	block.destroyable = false
	block.passable = true
	block.shootable = true
	return block
}

func (b *Block) Row() int {
	return b.row
}

func (b *Block) Column() int {
	return b.column
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

func (b *Block) Shootable() bool {
	return b.shootable
}

func (b *Block) IsDestroyed() bool {
	return b.quadrants == emptyQuadrants
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

func (b *Block) ProcessCollision(bullet *Bullet, sb *Block) {
	fb := b
	if sb == nil {
		sb = fb
	}
	switch bullet.Direction() {
	case North:
		y := 0
		if !fb.quadrants[0][0] && !fb.quadrants[1][0] && !sb.quadrants[0][0] && !sb.quadrants[1][0] {
			y = 1
		}
		fb.quadrants[0][y], fb.quadrants[1][y], sb.quadrants[0][y], sb.quadrants[1][y] = false, false, false, false
	case East:
		x := 0
		if !fb.quadrants[0][0] && !fb.quadrants[0][1] && !sb.quadrants[0][0] && !sb.quadrants[0][1] {
			x = 1
		}
		fb.quadrants[x][0], fb.quadrants[x][1], sb.quadrants[x][0], sb.quadrants[x][1] = false, false, false, false
	case South:
		y := 1
		if !fb.quadrants[0][1] && !fb.quadrants[1][1] && !sb.quadrants[0][1] && !sb.quadrants[1][1] {
			y = 0
		}
		fb.quadrants[0][y], fb.quadrants[1][y], sb.quadrants[0][y], sb.quadrants[1][y] = false, false, false, false
	case West:
		x := 1
		if !fb.quadrants[1][0] && !fb.quadrants[1][1] && !sb.quadrants[1][0] && !sb.quadrants[1][1] {
			x = 0
		}
		fb.quadrants[x][0], fb.quadrants[x][1], sb.quadrants[x][0], sb.quadrants[x][1] = false, false, false, false
	}
}
