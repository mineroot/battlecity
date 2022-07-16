package game

import "github.com/faiface/pixel/pixelgl"

type MainMenuState struct {
	config StateConfig
}

func NewMainMenuState(config StateConfig) *MainMenuState {
	return &MainMenuState{config: config}
}

func (s *MainMenuState) Update(_ *pixelgl.Window, _ float64) State {
	return NewStageTitleState(s.config, 1)
}

func (s *MainMenuState) Draw(_ *pixelgl.Window, _ float64) {

}
