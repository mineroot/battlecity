package game

import (
	"embed"
	"fmt"
	"log"
	"math"
	"math/rand"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/google/uuid"
)

const (
	stageColumns = 30
	stageRows    = 30
)

type Stage struct {
	Blocks         [stageColumns][stageRows]*Block
	blockSprites   map[string]*pixel.Sprite
	quadrantCanvas *pixelgl.Canvas
	batch          *pixel.Batch
	needsRedraw    bool
	botsPool       []BotType
	botPoolIndex   int
}

func NewStage(spritesheet pixel.Picture, scale float64, stagesConfigs embed.FS, stageNum int) *Stage {
	bytes, err := stagesConfigs.ReadFile(fmt.Sprintf("assets/stages/%d.stage", stageNum))
	if err != nil {
		panic(err)
	}

	data := string(bytes)
	// maxStageChars + new line chars
	maxChars := stageColumns*stageRows + stageColumns
	if len(data) != maxChars {
		log.Fatalf("field: invalid stage file length: %d", len(data))
	}
	var blocks [stageColumns][stageRows]*Block
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
	stage.batch = pixel.NewBatch(&pixel.TrianglesData{}, spritesheet)
	stage.blockSprites = map[string]*pixel.Sprite{
		BorderBlock: pixel.NewSprite(spritesheet, pixel.R(368, 248, 376, 256)),
		BrickBlock:  pixel.NewSprite(spritesheet, pixel.R(256, 184, 264, 192)),
		SteelBlock:  pixel.NewSprite(spritesheet, pixel.R(256, 176, 264, 184)),
		WaterBlock:  pixel.NewSprite(spritesheet, pixel.R(256, 192, 264, 200)),
	}
	stage.needsRedraw = true
	stage.initBotsPool(stageNum)
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

func (s *Stage) CreateBot(tanks map[uuid.UUID]Tank, spritesheet pixel.Picture) *Bot {
	for {
		randomColumn := float64(rand.Intn(27-3) + 3)
		newBotPos := pixel.V(randomColumn*BlockSize*Scale, 27*BlockSize*Scale)
		newBotRect := Rect(newBotPos, TankSize, TankSize)
		noIntersection := true
		for _, tank := range tanks {
			tankRect := Rect(tank.Pos(), TankSize, TankSize)
			intersect := tankRect.Intersect(newBotRect)
			if intersect != pixel.ZR {
				noIntersection = false
				break
			}
		}
		if noIntersection {
			if s.IsPoolEmpty() {
				return nil
			}
			botType := s.botsPool[s.botPoolIndex]
			s.botPoolIndex++
			return NewBot(spritesheet, BotType(botType), newBotPos)
		}
	}
}

func (s *Stage) initBotsPool(stageNum int) {
	// probability density function
	var pdf [4]float64
	avgBotsCount := 20
	switch stageNum {
	case 1:
		pdf = [4]float64{0.88, 0.12, 0, 0}
		avgBotsCount = 18
	case 2:
		pdf = [4]float64{0.7, 0.2, 0, 0.1}
	case 3:
		pdf = [4]float64{0.7, 0.2, 0, 0.1}
	case 4:
		pdf = [4]float64{0.1, 0.25, 0.5, 0.15}
	default:
		pdf = [4]float64{0.4, 0.25, 0.25, 0.1}
	}
	avgBotsCountDiff := rand.Intn(5) - 2 // [-2; 2]
	botsCount := avgBotsCount + avgBotsCountDiff

	// cumulative distribution function
	cdf := make([]float64, 4)
	cdf[0] = pdf[0]
	for i := 1; i < 4; i++ {
		cdf[i] = cdf[i-1] + pdf[i]
	}

	for i := 0; i < botsCount; i++ {
		botType := DefaultBot
		r := rand.Float64()
		for r > cdf[botType] {
			botType++
		}
		s.botsPool = append(s.botsPool, botType)
	}
}

func (s *Stage) IsPoolEmpty() bool {
	return s.botPoolIndex >= len(s.botsPool)
}
