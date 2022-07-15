package game

import (
	"embed"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

type State interface {
	Update(win *pixelgl.Window, dt float64) State
	Draw(win *pixelgl.Window, dt float64)
}

type StateConfig struct {
	Spritesheet   pixel.Picture
	StagesConfigs embed.FS
}

type Game struct {
	currentState State
}

func NewGame(config StateConfig) *Game {
	game := new(Game)
	game.currentState = NewMainMenuState(config)
	return game
}

func (g *Game) Run(win *pixelgl.Window, dt float64) {
	newState := g.currentState.Update(win, dt)
	if newState != nil {
		g.currentState = newState
		return
	}

	g.currentState.Draw(win, dt)
}
