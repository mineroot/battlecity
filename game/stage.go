package game

import (
	"embed"
	"fmt"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/google/uuid"
	"log"
	"math"
	"math/rand"
)

const (
	stageColumns = 30
	stageRows    = 30
)

type Stage struct {
	Blocks            [stageColumns][stageRows]*Block
	blockSprites      map[string]*pixel.Sprite
	hqSprite          *pixel.Sprite
	destroyedHQSprite *pixel.Sprite
	quadrantCanvas    *pixelgl.Canvas
	batch             *pixel.Batch
	needsRedraw       bool
	botsPool          []BotType
	botPoolIndex      int
	isHQArmored       bool
	isHQDestroyed     bool
}

func NewStage(spritesheet pixel.Picture, stagesConfigs embed.FS, stageNum int) *Stage {
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

		shiftX, shiftY := BlockSize*Scale/2, BlockSize*Scale/2
		x, y := float64(column)*BlockSize*Scale+shiftX, float64(30-row)*BlockSize*Scale-shiftY
		pos := pixel.V(x, y)

		switch blockSymbol {
		case BorderBlock:
			block = Border(pos, row, column)
		case BrickBlock:
			block = Brick(pos, row, column, shiftX, shiftY)
		case SteelBlock:
			block = Steel(pos, row, column)
		case WaterBlock:
			block = Water(pos, row, column)
		case HQBlock:
			block = HQ(pos, row, column)
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
	stage.hqSprite = pixel.NewSprite(spritesheet, pixel.R(304, 208, 320, 224))
	stage.destroyedHQSprite = pixel.NewSprite(spritesheet, pixel.R(320, 208, 336, 224))
	stage.needsRedraw = true
	stage.initBotsPool(stageNum)
	return stage
}

func (s *Stage) NeedsRedraw() {
	s.needsRedraw = true
}

func (s *Stage) ArmorHQ() {
	for _, hqArmorIndex := range s.getHQArmorIndexes() {
		row := hqArmorIndex[0]
		column := hqArmorIndex[1]
		block := s.Blocks[row][column]
		s.Blocks[row][column] = Steel(block.pos, block.row, block.column)
	}
	s.needsRedraw = true
	s.isHQArmored = true
}

func (s *Stage) DisarmorHQ() {
	shiftX, shiftY := BlockSize*Scale/2, BlockSize*Scale/2
	for _, hqArmorIndex := range s.getHQArmorIndexes() {
		row := hqArmorIndex[0]
		column := hqArmorIndex[1]
		block := s.Blocks[row][column]
		s.Blocks[row][column] = Brick(block.pos, block.row, block.column, shiftX, shiftY)
	}
	s.needsRedraw = true
	s.isHQArmored = false
}

func (s *Stage) DestroyHQ() {
	for _, hqIndex := range s.getHQIndexes() {
		row := hqIndex[0]
		column := hqIndex[1]
		block := s.Blocks[row][column]
		s.Blocks[row][column] = Space(block.pos, block.row, block.column)
	}
	s.needsRedraw = true
	s.isHQDestroyed = true
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
	hqPos := pixel.V(15*Scale*BlockSize, 3*Scale*BlockSize)
	hqM := pixel.IM.Moved(hqPos).Scaled(hqPos, Scale)
	if s.isHQDestroyed {
		s.destroyedHQSprite.Draw(s.batch, hqM)
	} else {
		s.hqSprite.Draw(s.batch, hqM)
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
			isBonus := false
			botType := s.botsPool[s.botPoolIndex]
			if s.botPoolIndex == 3 || s.botPoolIndex == 10 || s.botPoolIndex == len(s.botsPool)-3 {
				isBonus = true
			}
			s.botPoolIndex++
			return NewBot(spritesheet, botType, newBotPos, isBonus)
		}
	}
}

func (s *Stage) IsPoolEmpty() bool {
	return s.botPoolIndex >= len(s.botsPool)
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

func (s *Stage) getHQArmorIndexes() [8][2]int {
	return [8][2]int{
		{25, 13},
		{25, 14},
		{25, 15},
		{25, 16},
		{26, 13},
		{26, 16},
		{27, 13},
		{27, 16},
	}
}

func (s *Stage) getHQIndexes() [4][2]int {
	return [4][2]int{
		{26, 14},
		{26, 15},
		{27, 14},
		{27, 15},
	}
}
