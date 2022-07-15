package game

import (
	"math/rand"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/google/uuid"
	"golang.org/x/image/colornames"
)

type PlaygroundState struct {
	config           StateConfig
	currentStage     string
	stage            *Stage
	stageLoaded      bool
	player           *Player
	bots             []*Bot
	bullets          []*Bullet
	bulletSprite     *pixel.Sprite
	newBotInterval   time.Duration
	lastBotCreatedAt time.Time
}

func NewPlaygroundState(config StateConfig) *PlaygroundState {
	s := new(PlaygroundState)
	s.config = config
	s.currentStage = "1"
	s.bulletSprite = pixel.NewSprite(s.config.Spritesheet, pixel.R(323, 154, 326, 150))
	s.player = NewPlayer(s.config.Spritesheet)
	s.newBotInterval = time.Second * 3
	return s
}

func (s *PlaygroundState) Update(win *pixelgl.Window, dt float64) State {
	now := time.Now()
	if !s.stageLoaded {
		s.stage = NewStage(s.config.Spritesheet, Scale, s.config.StagesConfigs, s.currentStage)
		s.stageLoaded = true
	}

	// handle bots creation & movement
	canCreate := now.Sub(s.lastBotCreatedAt) > s.newBotInterval
	if len(s.bots) < 4 && canCreate {
		s.lastBotCreatedAt = now
		randomColumn := float64(rand.Intn(27-3) + 3)
		s.bots = append(s.bots, NewBot(s.config.Spritesheet, pixel.V(randomColumn*BlockSize*Scale, 27*BlockSize*Scale)))
	}

	// handle *all* tanks movement
	movementResults := make(map[uuid.UUID]*MovementResult)
	tanks := s.Tanks()
	for id, tank := range tanks {
		newPos, newDirection := tank.HandleMovement(win, dt)
		movementResults[id] = &MovementResult{newPos: newPos, direction: newDirection, canMove: true}
	}
	for _, blocks := range s.stage.Blocks {
		for _, block := range blocks {
			if !block.passable {
				blockRect := Rect(block.pos, BlockSize, BlockSize)
				for id, tank := range tanks {
					movementRes := movementResults[id]
					if tank.Pos() == movementRes.newPos { // tank didn't move
						continue
					}
					tankRect := Rect(movementRes.newPos, TankSize, TankSize)
					intersect := tankRect.Intersect(blockRect)
					if intersect != pixel.ZR { // collision detected
						movementRes.canMove = false
					}
				}

			}
		}
	}
	for id, tank := range tanks {
		tank.Move(movementResults[id], dt)
	}

	//newPos, newDirection := s.player.HandleMovement(win, dt)
	//if newPos != s.player.Pos() {
	//	playerCanMove := true
	//	playerRect := Rect(newPos, TankSize, TankSize)
	//	for _, blocks := range s.stage.Blocks {
	//		for _, block := range blocks {
	//			if !block.passable {
	//				blockRect := Rect(block.pos, BlockSize, BlockSize)
	//				intersect := playerRect.Intersect(blockRect)
	//				if intersect != pixel.ZR { // collision detected
	//					playerCanMove = false
	//				}
	//			}
	//		}
	//	}
	//
	//	s.player.Move(playerCanMove, newPos, newDirection)
	//}

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

func (s *PlaygroundState) Tanks() map[uuid.UUID]Tank {
	tanks := make(map[uuid.UUID]Tank)
	tanks[s.player.Id.Id()] = s.player
	for _, b := range s.bots {
		tanks[b.Id.Id()] = b
	}
	return tanks
}

func (s *PlaygroundState) Draw(win *pixelgl.Window, dt float64) {
	win.Clear(colornames.Black)
	s.stage.Draw(win)
	s.player.Draw(win, dt)
	for _, b := range s.bots {
		b.Draw(win, dt)
	}
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
