package entity

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
	spritesheet    pixel.Picture
	blockSprites   map[string]*pixel.Sprite
	quadrantCanvas *pixelgl.Canvas
	scale          float64
}

func NewStage(spritesheet pixel.Picture, scale float64, stagesConfigs embed.FS, stageNum int) *Stage {
	bytes, err := stagesConfigs.ReadFile(fmt.Sprintf("assets/stages/%d.stage", stageNum))
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

		w, h := 8.0, 8.0
		shiftX, shiftY := w*scale/2, h*scale/2
		x, y := float64(column)*w*scale+shiftX, float64(30-row)*h*scale-shiftY
		pos := pixel.V(x, y)

		switch blockSymbol {
		case BorderBlock:
			block = Border(pos)
		case BrickBlock:
			block = Brick(pos)
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
			block = Steel(pos)
		case WaterBlock:
			block = Water(pos)
		case SpaceBlock:
			block = Space(pos)
		default:
			log.Fatalf("field: invalid block symbol: %s", blockSymbol)
		}
		blocks[row][column] = block
		n++
	}

	stage := new(Stage)
	stage.Blocks = blocks
	stage.spritesheet = spritesheet
	stage.blockSprites = map[string]*pixel.Sprite{
		BorderBlock: pixel.NewSprite(stage.spritesheet, pixel.R(368, 248, 376, 256)),
		BrickBlock:  pixel.NewSprite(stage.spritesheet, pixel.R(256, 184, 264, 192)),
		SteelBlock:  pixel.NewSprite(stage.spritesheet, pixel.R(256, 176, 264, 184)),
		WaterBlock:  pixel.NewSprite(stage.spritesheet, pixel.R(256, 192, 264, 200)),
	}
	//stage.quadrantCanvas = pixelgl.NewCanvas(pixel.ZR)
	//stage.quadrantCanvas.Clear(colornames.Black)
	stage.scale = scale

	return stage
}

func (s *Stage) Draw(win *pixelgl.Window, dt float64) {
	for _, blocks := range s.Blocks {
		for _, block := range blocks {
			if sprite, ok := s.blockSprites[block.Kind()]; ok {
				sprite.Draw(win, pixel.IM.Moved(block.Pos()).Scaled(block.Pos(), s.scale))
				if block.destroyable {
					for i := 0; i < 2; i++ {
						for j := 0; j < 2; j++ {
							imd := block.QuadrantIMDraw(i, j)
							if imd != nil {
								imd.Draw(win)
							}
						}
					}
				}
			}
		}
	}
}
