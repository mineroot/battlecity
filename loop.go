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
	ml.spritesheet = spritesheet
	ml.stagesConfigs = stagesConfigs
	ml.player = entity.NewPlayer(ml.spritesheet)
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
		playerRect := Rect(newPos, entity.PlayerSize, entity.PlayerSize)
	outPlayer:
		for _, blocks := range ml.stage.Blocks {
			for _, block := range blocks {
				if !block.Passable() {
					blockRect := Rect(block.Pos(), entity.BlockSize, entity.BlockSize)
					intersect := playerRect.Intersect(blockRect)
					if intersect != pixel.ZR { // collision detected
						playerCanMove = false
						break outPlayer
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
	for i := 0; i < len(ml.bullets); i++ {
		bullet := ml.bullets[i]
		bullet.Move(dt)

		w, h := entity.BulletW, entity.BulletH
		if bullet.Direction().IsHorizontal() {
			w, h = h, w
		}
		bulletRect := Rect(bullet.Pos(), w, h)
		var collidedBlocks []*entity.Block
		for _, blocks := range ml.stage.Blocks {
			for _, block := range blocks {
				if !block.Shootable() {
					blockRect := Rect(block.Pos(), entity.BlockSize, entity.BlockSize)
					intersect := bulletRect.Intersect(blockRect)
					if intersect != pixel.ZR { // collision detected
						collidedBlocks = append(collidedBlocks, block)
					}
				}
			}
		}

		if len(collidedBlocks) != 0 {
			if len(collidedBlocks) > 2 {
				panic("theoretically impossible")
			}

			firstCollidedBlock := collidedBlocks[0]
			var secondCollidedBlock *entity.Block = nil
			if len(collidedBlocks) == 2 {
				secondCollidedBlock = collidedBlocks[1]
			}
			firstCollidedBlock.ProcessCollision(bullet, secondCollidedBlock)
			if firstCollidedBlock.IsDestroyed() {
				ml.stage.Destroy(firstCollidedBlock)
			}
			if secondCollidedBlock != nil && secondCollidedBlock.IsDestroyed() {
				ml.stage.Destroy(secondCollidedBlock)
			}

			// remove bullet
			ml.bullets[i] = ml.bullets[len(ml.bullets)-1]
			ml.bullets = ml.bullets[:len(ml.bullets)-1]
		}
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
	ml.stage.Draw(win)
	win.Update()
}

func (ml *MainLoop) DrawBullets(win *pixelgl.Window) {
	for _, bullet := range ml.bullets {
		m := pixel.IM.Moved(bullet.Pos()).
			Scaled(bullet.Pos(), entity.Scale).
			Rotated(bullet.Pos(), bullet.Direction().Angle())
		ml.bulletSprite.Draw(win, m)
	}
}

func Rect(pos pixel.Vec, w float64, h float64) pixel.Rect {
	w, h = w*entity.Scale/2, h*entity.Scale/2
	return pixel.R(pos.X-w, pos.Y-h, pos.X+w, pos.Y+h)
}

func (ml *MainLoop) loadCurrentStage() {
	ml.stage = entity.NewStage(ml.spritesheet, entity.Scale, ml.stagesConfigs, ml.currentStage)
}
