package game

import (
	"battlecity/game/utils"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
)

const (
	BorderBlock = "|"
	BrickBlock  = "b"
	SteelBlock  = "s"
	WaterBlock  = "w"
	HQBlock     = "h"
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
	destroyable    bool // can Bullet destroy it
	passable       bool // can Tank pass through it
	shootable      bool // can Bullet pass through it
	bonus          bool // can Bonus appears on it
	pos            pixel.Vec
	quadrants      [2][2]bool
	quadrantIMDraw [2][2]*imdraw.IMDraw
}

func Border(pos pixel.Vec, row, column int) *Block {
	block := new(Block)
	block.row, block.column, block.pos = row, column, pos
	block.kind = BorderBlock
	block.destroyable = false
	block.passable = false
	block.shootable = false
	block.bonus = false
	return block
}

func Brick(pos pixel.Vec, row, column int, shiftX, shiftY float64) *Block {
	block := new(Block)
	block.row, block.column, block.pos = row, column, pos
	block.kind = BrickBlock
	block.destroyable = true
	block.passable = false
	block.shootable = false
	block.bonus = true
	block.quadrants = fullQuadrants
	block.InitQuadrants(shiftX, shiftY)
	return block
}

func Steel(pos pixel.Vec, row, column int) *Block {
	block := new(Block)
	block.row, block.column, block.pos = row, column, pos
	block.kind = SteelBlock
	block.destroyable = false
	block.passable = false
	block.shootable = false
	block.bonus = false
	return block
}

func Water(pos pixel.Vec, row, column int) *Block {
	block := new(Block)
	block.row, block.column, block.pos = row, column, pos
	block.kind = WaterBlock
	block.destroyable = false
	block.passable = false
	block.shootable = true
	block.bonus = false
	return block
}

func HQ(pos pixel.Vec, row, column int) *Block {
	block := new(Block)
	block.row, block.column = row, column
	block.pos = pos
	block.kind = HQBlock
	block.destroyable = true
	block.passable = false
	block.shootable = false
	block.bonus = false
	return block
}

func Space(pos pixel.Vec, row, column int) *Block {
	block := new(Block)
	block.row, block.column, block.pos = row, column, pos
	block.kind = SpaceBlock
	block.destroyable = false
	block.passable = true
	block.shootable = true
	block.bonus = true
	return block
}

func (b *Block) IsDestroyed() bool {
	return b.quadrants == emptyQuadrants
}

func (b *Block) InitQuadrants(shiftX, shiftY float64) {
	if !b.destroyable {
		return
	}
	var quadrantRects [2][2]pixel.Rect
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			r := pixel.R(b.pos.X-shiftX, b.pos.Y-shiftY, b.pos.X, b.pos.Y)
			r = r.Moved(pixel.V(shiftX*float64(i), shiftY*float64(j)))
			quadrantRects[i][j] = r
		}
	}
	for i, rectsI := range quadrantRects {
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
	switch bullet.direction {
	case utils.North:
		y := 0
		if !fb.quadrants[0][0] && !fb.quadrants[1][0] && !sb.quadrants[0][0] && !sb.quadrants[1][0] {
			y = 1
		}
		fb.quadrants[0][y], fb.quadrants[1][y], sb.quadrants[0][y], sb.quadrants[1][y] = false, false, false, false
	case utils.East:
		x := 0
		if !fb.quadrants[0][0] && !fb.quadrants[0][1] && !sb.quadrants[0][0] && !sb.quadrants[0][1] {
			x = 1
		}
		fb.quadrants[x][0], fb.quadrants[x][1], sb.quadrants[x][0], sb.quadrants[x][1] = false, false, false, false
	case utils.South:
		y := 1
		if !fb.quadrants[0][1] && !fb.quadrants[1][1] && !sb.quadrants[0][1] && !sb.quadrants[1][1] {
			y = 0
		}
		fb.quadrants[0][y], fb.quadrants[1][y], sb.quadrants[0][y], sb.quadrants[1][y] = false, false, false, false
	case utils.West:
		x := 1
		if !fb.quadrants[1][0] && !fb.quadrants[1][1] && !sb.quadrants[1][0] && !sb.quadrants[1][1] {
			x = 0
		}
		fb.quadrants[x][0], fb.quadrants[x][1], sb.quadrants[x][0], sb.quadrants[x][1] = false, false, false, false
	}
}
