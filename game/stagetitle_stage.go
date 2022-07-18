package game

import (
	"fmt"
	"image/color"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"

	"golang.org/x/image/colornames"
)

type StageTitleState struct {
	config         StateConfig
	stageNum       int
	stateStartTime time.Time
	stageTxt       *text.Text
	player         *Player
}

func NewStageTitleState(config StateConfig, stageNum int, player *Player) *StageTitleState {
	s := new(StageTitleState)
	s.config = config
	s.stageNum = stageNum
	s.player = player

	atlas := text.NewAtlas(s.config.DefaultFont, text.ASCII)
	s.stageTxt = text.New(pixel.V(0, 0), atlas)
	s.stageTxt.Color = colornames.Black
	var txt string
	if s.stageNum < 10 {
		txt = fmt.Sprintf("STAGE  %d", s.stageNum)
	} else {
		txt = fmt.Sprintf("STAGE %d", s.stageNum)
	}
	r := s.stageTxt.BoundsOf(txt)
	s.stageTxt.Orig.X = config.WindowBounds.W()/2 - r.W()/2
	s.stageTxt.Orig.Y = config.WindowBounds.H()/2 - r.H()/2
	_, _ = fmt.Fprintln(s.stageTxt, txt)

	return s
}

func (s *StageTitleState) Update(_ *pixelgl.Window, _ float64) State {
	now := time.Now()
	if s.stateStartTime.IsZero() {
		s.stateStartTime = now
	}
	if now.Sub(s.stateStartTime) >= time.Second*3 {
		return NewPlaygroundState(s.config, s.stageNum, s.player)
	}
	return nil
}

func (s *StageTitleState) Draw(win *pixelgl.Window, _ float64) {
	win.Clear(color.RGBA{R: 99, G: 99, B: 99, A: 1})
	s.stageTxt.Draw(win, pixel.IM)
}
