package game

import (
	"embed"
	"fmt"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"log"
	"math"
)

type Stage struct {
	Blocks         [30][30]*Block
	blockSprites   map[string]*pixel.Sprite
	quadrantCanvas *pixelgl.Canvas
	batch          *pixel.Batch
	needsRedraw    bool
}

func NewStage(spritesheet *pixel.Picture, scale float64, stagesConfigs embed.FS, stageName string) *Stage {
	bytes, err := stagesConfigs.ReadFile(fmt.Sprintf("assets/stages/%s.stage", stageName))
	if err != nil {
		panic(err)
	}

	data := string(bytes)
	if len(data) != (30*31 - 1) {
		log.Fatalf("field: invalid stage file length: %d", len(data))
	}
	var blocks [30][30]*Block
	var block *Block
	n := 0
	for _, ch := range data {
		blockSymbol := string(ch)
		if blockSymbol == "\n" {
			continue
		}

		row := n / 30
		column := int(math.Mod(float64(n), 30))

		shiftX, shiftY := BlockSize*scale/2, BlockSize*scale/2
		x, y := float64(column)*BlockSize*scale+shiftX, float64(30-row)*BlockSize*scale-shiftY
		pos := pixel.V(x, y)

		switch blockSymbol {
		case BorderBlock:
			block = Border(pos, row, column)
		case BrickBlock:
			block = Brick(pos, row, column)
			var quadrantRects [2][2]pixel.Rect
			for i := 0; i < 2; i++ {
				for j := 0; j < 2; j++ {
					r := pixel.R(pos.X-shiftX, pos.Y-shiftY, pos.X, pos.Y)
					r = r.Moved(pixel.V(shiftX*float64(i), shiftY*float64(j)))
					quadrantRects[i][j] = r
				}
			}
			block.InitQuadrants(quadrantRects)
		case SteelBlock:
			block = Steel(pos, row, column)
		case WaterBlock:
			block = Water(pos, row, column)
		case SpaceBlock:
			block = Space(pos, row, column)
		default:
			log.Fatalf("field: invalid block symbol: %s", blockSymbol)
		}
		blocks[row][column] = block
		n++
	}

	stage := new(Stage)
	stage.Blocks = blocks
	stage.batch = pixel.NewBatch(&pixel.TrianglesData{}, *spritesheet)
	stage.blockSprites = map[string]*pixel.Sprite{
		BorderBlock: pixel.NewSprite(*spritesheet, pixel.R(368, 248, 376, 256)),
		BrickBlock:  pixel.NewSprite(*spritesheet, pixel.R(256, 184, 264, 192)),
		SteelBlock:  pixel.NewSprite(*spritesheet, pixel.R(256, 176, 264, 184)),
		WaterBlock:  pixel.NewSprite(*spritesheet, pixel.R(256, 192, 264, 200)),
	}
	stage.needsRedraw = true
	return stage
}

func (s *Stage) NeedsRedraw() {
	s.needsRedraw = true
}

func (s *Stage) DestroyBlock(block *Block) {
	s.Blocks[block.row][block.column] = Space(block.pos, block.row, block.column)
	s.needsRedraw = true
}

func (s *Stage) Draw(win *pixelgl.Window) {
	if !s.needsRedraw {
		s.batch.Draw(win)
		return
	}

	s.batch.Clear()
	for _, blocks := range s.Blocks {
		for _, block := range blocks {
			if sprite, ok := s.blockSprites[block.kind]; ok {
				sprite.Draw(s.batch, pixel.IM.Moved(block.pos).Scaled(block.pos, Scale))
				if block.destroyable {
					for i := 0; i < 2; i++ {
						for j := 0; j < 2; j++ {
							imd := block.QuadrantIMDraw(i, j)
							if imd != nil {
								imd.Draw(s.batch)
							}
						}
					}
				}
			}
		}
	}
	s.batch.Draw(win)
	s.needsRedraw = false
}
