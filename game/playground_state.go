package game

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

type PlaygroundState struct {
	config       StateConfig
	currentStage string
	stage        *Stage
	stageLoaded  bool
	player       *Player
	bots         []*Bot
	bullets      []*Bullet
	bulletSprite *pixel.Sprite
}

func NewPlaygroundState(config StateConfig) *PlaygroundState {
	s := new(PlaygroundState)
	s.config = config
	s.currentStage = "1"
	s.bulletSprite = pixel.NewSprite(s.config.Spritesheet, pixel.R(323, 154, 326, 150))
	s.player = NewPlayer(s.config.Spritesheet)
	return s
}

func (s *PlaygroundState) Update(win *pixelgl.Window, dt float64) State {
	if !s.stageLoaded {
		s.stage = NewStage(s.config.Spritesheet, Scale, s.config.StagesConfigs, s.currentStage)
		s.stageLoaded = true
	}

	// handle player movement
	//playerDt := 0.0
	newPos, newDirection := s.player.HandleMovementInput(win, dt)
	if newPos != s.player.Pos() {
		//playerDt = dt
		playerCanMove := true
		playerRect := Rect(newPos, PlayerSize, PlayerSize)
	outPlayer:
		for _, blocks := range s.stage.Blocks {
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

		s.player.Move(playerCanMove, newPos, newDirection)
	}

	// handle bullets movement
	for i := 0; i < len(s.bullets); i++ {
		bullet := s.bullets[i]
		bullet.Move(dt)

		w, h := BulletW, BulletH
		if bullet.direction.IsHorizontal() {
			w, h = h, w
		}
		bulletRect := Rect(bullet.pos, w, h)
		var collidedDestroyableBlocks []*Block
		collision := false
		for _, blocks := range s.stage.Blocks {
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
				s.stage.DestroyBlock(firstCollidedBlock)
			}
			if secondCollidedBlock != nil && secondCollidedBlock.IsDestroyed() {
				s.stage.DestroyBlock(secondCollidedBlock)
			}
			s.stage.NeedsRedraw()
		}
		// remove bullet
		if collision {
			bullet.Destroy()
			s.bullets[i] = s.bullets[len(s.bullets)-1]
			s.bullets = s.bullets[:len(s.bullets)-1]
		}
	}

	// handle shooting input
	playerBullet := s.player.HandleShootingInput(win)
	if playerBullet != nil {
		s.bullets = append(s.bullets, playerBullet)
	}

	return nil
}

func (s *PlaygroundState) Draw(win *pixelgl.Window, dt float64) {
	win.Clear(colornames.Black)
	s.stage.Draw(win)
	s.player.Draw(win, dt)
	s.DrawBullets(win)
}

func (s *PlaygroundState) DrawBullets(win *pixelgl.Window) {
	for _, bullet := range s.bullets {
		m := pixel.IM.Moved(bullet.pos).
			Scaled(bullet.pos, Scale).
			Rotated(bullet.pos, bullet.direction.Angle())
		s.bulletSprite.Draw(win, m)
	}
}
