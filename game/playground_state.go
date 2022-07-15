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
	bots             map[uuid.UUID]*Bot
	bullets          []*Bullet
	bulletSprite     *pixel.Sprite
	newBotInterval   time.Duration
	lastBotCreatedAt time.Time
	isPaused         bool
}

func NewPlaygroundState(config StateConfig) *PlaygroundState {
	s := new(PlaygroundState)
	s.config = config
	s.currentStage = "1"
	s.bulletSprite = pixel.NewSprite(s.config.Spritesheet, pixel.R(323, 154, 326, 150))
	s.player = NewPlayer(s.config.Spritesheet)
	s.bots = make(map[uuid.UUID]*Bot)
	s.newBotInterval = time.Second * 3
	return s
}

func (s *PlaygroundState) Update(win *pixelgl.Window, dt float64) State {
	if win.JustPressed(pixelgl.KeyEscape) {
		s.isPaused = !s.isPaused
	}
	if s.isPaused {
		return nil
	}
	const maxBots = 4
	tanks := s.Tanks()
	now := time.Now()
	if !s.stageLoaded {
		s.stage = NewStage(s.config.Spritesheet, Scale, s.config.StagesConfigs, s.currentStage)
		s.stageLoaded = true
	}

	// handle bots creation
	canCreate := now.Sub(s.lastBotCreatedAt) > s.newBotInterval
	if len(s.bots) < maxBots && canCreate {
		s.lastBotCreatedAt = now
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
				}
			}
			if noIntersection {
				botType := rand.Intn(4)
				newBot := NewBot(s.config.Spritesheet, BotType(botType), newBotPos)
				s.bots[newBot.id] = newBot
				break
			}
		}
	}

	// handle *all* tanks movement
	movementResults := make(map[uuid.UUID]*MovementResult)
	for id, tank := range tanks {
		newPos, newDirection := tank.CalculateMovement(win, dt)
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
	for idI := range tanks {
		movementResultI := movementResults[idI]
		if !movementResultI.canMove { // already can't move - skip
			continue
		}
		tankIRect := Rect(movementResultI.newPos, TankSize, TankSize)
		for idJ, tankJ := range tanks {
			if idI == idJ { // don't compare with itself - skip
				continue
			}
			//movementResultJ := movementResults[idJ]
			tankJRect := Rect(tankJ.Pos(), TankSize, TankSize)

			intersect := tankIRect.Intersect(tankJRect)
			if intersect != pixel.ZR { // collision detected
				movementResultI.canMove = false
			}
		}
	}

	for id, tank := range tanks {
		tank.Move(movementResults[id], dt)
	}

	// handle bullets movement & collision
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
		for _, blocks := range s.stage.Blocks { // check collision between bullet and blocks
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
		} else { // check collision between bullet and tanks
			for id, tank := range tanks {
				if tank.Side() != bullet.origin.Side() {
					tankRect := Rect(tank.Pos(), TankSize, TankSize)
					intersect := bulletRect.Intersect(tankRect)
					if intersect != pixel.ZR { // collision detected
						if tank.Side() == bot {
							botTank, _ := tank.(*Bot)
							botTank.hp--
							if botTank.hp <= 0 {
								delete(s.bots, id)
							}
						} else {
							// TODO: decrease player *hp* points
						}
						collision = true
					}
				}
			}
		}
		// check bullets collision
		if !collision {
			for j := 0; j < len(s.bullets); j++ {
				bullet2 := s.bullets[j]
				if i != j && bullet.origin.Side() != bullet2.origin.Side() {
					w2, h2 := BulletW, BulletH
					if bullet.direction.IsHorizontal() {
						w2, h2 = h2, w2
					}
					bullet2Rect := Rect(bullet2.pos, w2, h2)
					intersect := bulletRect.Intersect(bullet2Rect)
					if intersect != pixel.ZR { // collision detected
						collision = true
						bullet2.Destroy()
					}
				}
			}
		}

		if collision {
			bullet.Destroy()
		}
		// remove destroyed bullets from slice
		var tmpBullets []*Bullet
		for _, b := range s.bullets {
			if !b.destroyed {
				tmpBullets = append(tmpBullets, b)
			}
		}
		s.bullets = tmpBullets

	}

	// handle shooting
	for _, tank := range tanks {
		bullet := tank.Shoot(win, dt)
		if bullet != nil {
			s.bullets = append(s.bullets, bullet)
		}
	}

	return nil
}

func (s *PlaygroundState) Tanks() map[uuid.UUID]Tank {
	tanks := make(map[uuid.UUID]Tank)
	tanks[s.player.id] = s.player
	for id, b := range s.bots {
		tanks[id] = b
	}
	return tanks
}

func (s *PlaygroundState) Draw(win *pixelgl.Window, dt float64) {
	if s.isPaused {
		dt = 0
	}
	win.Clear(colornames.Black)
	s.stage.Draw(win)
	s.DrawBullets(win)
	s.player.Draw(win, dt)
	for _, b := range s.bots {
		b.Draw(win, dt)
	}
}

func (s *PlaygroundState) DrawBullets(win *pixelgl.Window) {
	for _, bullet := range s.bullets {
		m := pixel.IM.Moved(bullet.pos).
			Scaled(bullet.pos, Scale).
			Rotated(bullet.pos, bullet.direction.Angle())
		s.bulletSprite.Draw(win, m)
	}
}
