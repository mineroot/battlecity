package game

import (
	"embed"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

type MainLoop struct {
	currentStage  string
	spritesheet   *pixel.Picture
	stagesConfigs embed.FS
	player        *Player
	stage         *Stage
	bullets       []*Bullet
	bulletSprite  *pixel.Sprite
}

func CreateMainLoop(spritesheet *pixel.Picture, stagesConfigs embed.FS) *MainLoop {
	ml := new(MainLoop)
	ml.currentStage = "1"
	ml.spritesheet = spritesheet
	ml.stagesConfigs = stagesConfigs
	ml.player = NewPlayer(*ml.spritesheet)
	ml.bulletSprite = pixel.NewSprite(*ml.spritesheet, pixel.R(323, 154, 326, 150))
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
		playerRect := Rect(newPos, PlayerSize, PlayerSize)
	outPlayer:
		for _, blocks := range ml.stage.Blocks {
			for _, block := range blocks {
				if !block.passable {
					blockRect := Rect(block.pos, BlockSize, BlockSize)
					intersect := playerRect.Intersect(blockRect)
					if intersect != pixel.ZR { // collision detected
						playerCanMove = false
						break outPlayer
					}
				}
			}
		}

		ml.player.Move(playerCanMove, newPos, newDirection)
	}

	// handle bullets movement
	for i := 0; i < len(ml.bullets); i++ {
		bullet := ml.bullets[i]
		bullet.Move(dt)

		w, h := BulletW, BulletH
		if bullet.direction.IsHorizontal() {
			w, h = h, w
		}
		bulletRect := Rect(bullet.pos, w, h)
		var collidedDestroyableBlocks []*Block
		collision := false
		for _, blocks := range ml.stage.Blocks {
			for _, block := range blocks {
				if !block.shootable {
					blockRect := Rect(block.pos, BlockSize, BlockSize)
					intersect := bulletRect.Intersect(blockRect)
					if intersect != pixel.ZR { // collision detected
						if block.destroyable {
							collidedDestroyableBlocks = append(collidedDestroyableBlocks, block)
						}
						collision = true
					}
				}
			}
		}

		if len(collidedDestroyableBlocks) != 0 {
			if len(collidedDestroyableBlocks) > 2 {
				panic("theoretically impossible")
			}

			firstCollidedBlock := collidedDestroyableBlocks[0]
			var secondCollidedBlock *Block = nil
			if len(collidedDestroyableBlocks) == 2 {
				secondCollidedBlock = collidedDestroyableBlocks[1]
			}
			firstCollidedBlock.ProcessCollision(bullet, secondCollidedBlock)
			if firstCollidedBlock.IsDestroyed() {
				ml.stage.DestroyBlock(firstCollidedBlock)
			}
			if secondCollidedBlock != nil && secondCollidedBlock.IsDestroyed() {
				ml.stage.DestroyBlock(secondCollidedBlock)
			}
			ml.stage.NeedsRedraw()
		}
		// remove bullet
		if collision {
			bullet.Destroy()
			ml.bullets[i] = ml.bullets[len(ml.bullets)-1]
			ml.bullets = ml.bullets[:len(ml.bullets)-1]
		}
	}

	// handle shooting input
	playerBullet := ml.player.HandleShootingInput(win)
	if playerBullet != nil {
		ml.bullets = append(ml.bullets, playerBullet)
	}

	_ = playerDt
	// draw all
	win.Clear(colornames.Black)
	ml.stage.Draw(win)
	ml.player.Draw(win, playerDt)
	ml.DrawBullets(win)
	win.Update()
}

func (ml *MainLoop) DrawBullets(win *pixelgl.Window) {
	for _, bullet := range ml.bullets {
		m := pixel.IM.Moved(bullet.pos).
			Scaled(bullet.pos, Scale).
			Rotated(bullet.pos, bullet.direction.Angle())
		ml.bulletSprite.Draw(win, m)
	}
}

func Rect(pos pixel.Vec, w float64, h float64) pixel.Rect {
	w, h = w*Scale/2, h*Scale/2
	return pixel.R(pos.X-w, pos.Y-h, pos.X+w, pos.Y+h)
}

func (ml *MainLoop) loadCurrentStage() {
	ml.stage = NewStage(ml.spritesheet, Scale, ml.stagesConfigs, ml.currentStage)
}
