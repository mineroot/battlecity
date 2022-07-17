package game

import (
	"math/rand"
	"time"

	"battlecity/game/utils"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

type BonusType int

const BonusSize = 16

const (
	ImmunityBonus BonusType = iota
	TimeStopBonus
	HQArmorBonus
	UpgradeBonus
	AnnihilationBonus
	LifeBonus
)

type Bonus struct {
	pos       pixel.Vec
	bonusType BonusType
	model     *utils.Animation
}

func NewBonus(spritesheet pixel.Picture, blocks [stageColumns][stageRows]*Block) *Bonus {
	bonus := new(Bonus)
	bonus.bonusType = BonusType(rand.Intn(6))
	var pos pixel.Vec
	for {
		row, column := rand.Intn(stageRows), rand.Intn(stageColumns)
		block := blocks[row][column]
		if block.bonus {
			r := Rect(block.pos, BlockSize, BlockSize)
			minMaxPoints := []pixel.Vec{r.Min, r.Max}
			pos = minMaxPoints[rand.Intn(len(minMaxPoints))]
			break
		}
	}
	bonus.pos = pos
	duration := time.Millisecond * 150
	minXStart, maxXStart, frameW := 256, 272, 16
	minX := float64(minXStart + frameW*int(bonus.bonusType))
	maxX := float64(maxXStart + frameW*int(bonus.bonusType))
	bonus.model = utils.NewAnimation([]utils.AnimationFrame{
		{Frame: pixel.NewSprite(spritesheet, pixel.R(minX, 128, maxX, 144)), Duration: duration},
		{Frame: nil, Duration: duration},
	})

	return bonus
}

func (b *Bonus) Draw(win *pixelgl.Window, dt float64) {
	frame := b.model.CurrentFrame(dt)
	if frame != nil {
		frame.Draw(win, pixel.IM.Moved(b.pos).Scaled(b.pos, Scale))
	}
}
