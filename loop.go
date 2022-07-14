package main

import (
	"battlecity2/entity"
	"embed"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

type MainLoop struct {
	currentStage  int
	scale         float64
	spritesheet   pixel.Picture
	stagesConfigs embed.FS
	player        *entity.Player
	stage         *entity.Stage
	bullets       []*entity.Bullet
	bulletSprite  *pixel.Sprite
}

func CreateMainLoop(spritesheet pixel.Picture, stagesConfigs embed.FS) *MainLoop {
	ml := new(MainLoop)
	ml.currentStage = 1
	ml.scale = 4
	ml.spritesheet = spritesheet
	ml.stagesConfigs = stagesConfigs
	ml.player = entity.NewPlayer(ml.spritesheet, ml.scale)
	ml.bulletSprite = pixel.NewSprite(ml.spritesheet, pixel.R(323, 154, 326, 150))
	ml.loadCurrentStage()
	return ml
}

func (ml *MainLoop) Run(win *pixelgl.Window, dt float64) {
	// handle player movement
	playerDt := 0.0
	newPos, newDirection := ml.player.HandleMovementInput(win, dt)
	if newPos != ml.player.Pos() {
		playerDt = dt
		playerCanMove := true
		playerRect := Rect(newPos, entity.PlayerSize, entity.PlayerSize, ml.scale)
	out:
		for _, blocks := range ml.stage.Blocks {
			for _, block := range blocks {
				if !block.Passable() {
					blockRect := Rect(block.Pos(), 8, 8, ml.scale)
					intersect := playerRect.Intersect(blockRect)
					if intersect != pixel.ZR { // collision detected
						playerCanMove = false
						break out
					}
				}
			}
		}

		if !playerCanMove {
			newPos = ml.player.Pos()
		}
		ml.player.Move(newPos, newDirection)
	}

	// handle bullets movement
	for _, bullet := range ml.bullets {
		bullet.Move(dt)
	}

	// handle shooting input
	playerBullet := ml.player.HandleShootingInput(win)
	if playerBullet != nil {
		ml.bullets = append(ml.bullets, playerBullet)
	}

	// draw all
	win.Clear(colornames.Black)
	ml.DrawBullets(win)
	ml.player.Draw(win, playerDt)
	ml.stage.Draw(win, dt)
	win.Update()
}

func (ml *MainLoop) DrawBullets(win *pixelgl.Window) {
	for _, bullet := range ml.bullets {
		m := pixel.IM.Moved(bullet.Pos()).
			Scaled(bullet.Pos(), ml.scale).
			Rotated(bullet.Pos(), bullet.Direction().Angle())
		ml.bulletSprite.Draw(win, m)
	}
}

func Rect(pos pixel.Vec, w float64, h float64, scale float64) pixel.Rect {
	w, h = w*scale/2, h*scale/2
	return pixel.R(pos.X-w, pos.Y-h, pos.X+w, pos.Y+h)
}

func (ml *MainLoop) loadCurrentStage() {
	ml.stage = entity.NewStage(ml.spritesheet, ml.scale, ml.stagesConfigs, ml.currentStage)
}
