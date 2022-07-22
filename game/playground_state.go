package game

import (
	"battlecity/game/explosions"
	"battlecity/game/sfx"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/google/uuid"
	"golang.org/x/image/colornames"
	"math"
	"time"
)

type PlaygroundState struct {
	config            StateConfig
	stageNum          int
	rSide             *RSide
	stage             *Stage
	player            *Player
	bots              map[uuid.UUID]*Bot
	destroyedBots     []BotType
	activeBonus       *Bonus
	bullets           []*Bullet
	bulletSprite      *pixel.Sprite
	newBotInterval    time.Duration
	newBotDuration    time.Duration
	stageClearedTime  time.Time
	isPaused          bool
	isTimeStopBonus   bool
	timeStopDuration  time.Duration
	isArmoredHQBonus  bool
	armoredHQDuration time.Duration
	explosions        []*explosions.Explosion
}

func NewPlaygroundState(config StateConfig, stageNum int, player *Player) *PlaygroundState {
	s := new(PlaygroundState)
	s.config = config
	s.rSide = NewRightSide(s.config.Spritesheet, s.config.DefaultFont)
	s.stageNum = stageNum
	s.bulletSprite = pixel.NewSprite(s.config.Spritesheet, pixel.R(323, 154, 326, 150))
	if player == nil {
		s.player = NewPlayer(s.config.Spritesheet)
		s.player.ResetLevel()
	} else {
		s.player = player
	}
	s.player.Respawn()
	s.bots = make(map[uuid.UUID]*Bot)
	s.newBotInterval = time.Second * 3
	s.stage = NewStage(s.config.Spritesheet, s.config.StagesConfigs, s.stageNum)
	return s
}

func (s *PlaygroundState) Update(win *pixelgl.Window, dt float64) State {
	now := time.Now()
	if s.isStageCleared() && now.Sub(s.stageClearedTime) >= (time.Second*3) {
		return NewStageTitleState(s.config, s.stageNum+1, s.player)
	}
	if win.JustPressed(pixelgl.KeyEscape) && !s.isStageCleared() {
		s.isPaused = !s.isPaused
		if s.isPaused {
			sfx.PlayPause()
		} else {
			sfx.StopPause()
		}
	}
	if s.isPaused {
		return nil
	}

	const maxBots = 4
	tanks := s.Tanks()

	s.player.Update(dt)
	// handle bots creation
	canCreate := s.newBotDuration > s.newBotInterval || (len(s.destroyedBots) == 0 && len(s.bots) == 0)
	s.newBotDuration += time.Duration(dt * float64(time.Second))
	if len(s.bots) < maxBots && canCreate {
		if newBot := s.stage.CreateBot(tanks, s.config.Spritesheet); newBot != nil {
			s.bots[newBot.id] = newBot
			if newBot.isBonus {
				s.activeBonus = nil
			}
			s.newBotDuration = 0
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
			tankJRect := Rect(tankJ.Pos(), TankSize, TankSize)

			intersect := tankIRect.Intersect(tankJRect)
			if intersect != pixel.ZR { // collision detected
				movementResultI.canMove = false
			}
		}
	}

	if s.isTimeStopBonus {
		s.timeStopDuration += time.Duration(dt * float64(time.Second))
		if s.timeStopDuration > time.Second*10 {
			s.isTimeStopBonus = false
		}
	}
	if s.isArmoredHQBonus {
		if s.armoredHQDuration == 0 {
			s.stage.ArmorHQ()
		}
		s.armoredHQDuration += time.Duration(dt * float64(time.Second))
		if s.armoredHQDuration >= time.Second*17 {
			blinkPeriod := time.Millisecond * 250
			delta := s.armoredHQDuration - (time.Second * 17)
			if math.Mod(float64(delta/blinkPeriod), 2) == 0 {
				if s.stage.isHQArmored {
					s.stage.DisarmorHQ()
				}
			} else {
				if !s.stage.isHQArmored {
					s.stage.ArmorHQ()
				}
			}
		}
		if s.armoredHQDuration >= time.Second*20 {
			s.isArmoredHQBonus = false
		}
	}
	for id, tank := range tanks {
		canMove := tank.Side() == human || tank.Side() == bot && !s.isTimeStopBonus
		if canMove {
			tank.Move(movementResults[id], dt)
		}
	}

	s.bonusUpdate()

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
						if block.destroyable || (block.kind == SteelBlock && bullet.IsUpgraded()) {
							if block.kind == HQBlock {
								s.stage.DestroyHQ()
								sfx.PlayHQDestroyed()
								hqPos := pixel.V(15*Scale*BlockSize, 3*Scale*BlockSize)
								explosion := explosions.NewExplosion(explosions.TankExplosion, hqPos)
								s.explosions = append(s.explosions, explosion)
								// TODO game over
							} else {
								collidedDestroyableBlocks = append(collidedDestroyableBlocks, block)
							}
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
			if firstCollidedBlock.IsDestroyed() || bullet.IsUpgraded() {
				s.stage.DestroyBlock(firstCollidedBlock)
			}
			if secondCollidedBlock != nil && (secondCollidedBlock.IsDestroyed() || bullet.IsUpgraded()) {
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
							if botTank.isBonus {
								sfx.PlayBonusAppeared()
								botTank.isBonus = false
								s.activeBonus = NewBonus(s.config.Spritesheet, s.stage.Blocks)
							}
							if botTank.hp <= 0 {
								s.destroyBot(id)
							}
						} else if !s.player.immune {
							s.player.lives--
							if s.player.lives < 0 {
								// TODO game over
							}
							sfx.PlayPlayerDestroyed()
							explosion := explosions.NewExplosion(explosions.TankExplosion, s.player.pos)
							s.explosions = append(s.explosions, explosion)
							s.player.ResetLevel()
							s.player.Respawn()
						}
						collision = true
					}
				}
			}
		}
		// check collision between bullet and bullet
		isBulletBulletCollision := false
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
						isBulletBulletCollision = true
						bullet2.Destroy()
					}
				}
			}
		}

		if collision {
			if !isBulletBulletCollision {
				explosion := explosions.NewExplosion(explosions.BulletExplosion, bullet.pos)
				s.explosions = append(s.explosions, explosion)
			}
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

	// handle explosions
	for i := 0; i < len(s.explosions); i++ {
		exp := s.explosions[i]
		if exp.IsEnded() {
			s.explosions[i] = s.explosions[len(s.explosions)-1]
			s.explosions = s.explosions[:len(s.explosions)-1]
		}
	}

	// handle shooting
	for _, tank := range tanks {
		canShoot := tank.Side() == human || tank.Side() == bot && !s.isTimeStopBonus
		if canShoot {
			bullet := tank.Shoot(win, dt)
			if bullet != nil {
				s.bullets = append(s.bullets, bullet)
			}
		}
	}

	if s.isStageCleared() && s.stageClearedTime.IsZero() {
		s.stageClearedTime = now
	}
	s.rSide.Update(RSideData{
		stageNum:         s.stageNum,
		firstPlayerLives: int(math.Max(float64(s.player.lives), 0)),
		botsPullLen:      len(s.stage.botsPool) - s.stage.botPoolIndex,
	})

	return nil
}

func (s *PlaygroundState) Draw(win *pixelgl.Window, dt float64) {
	win.Clear(colornames.Black)
	s.stage.Draw(win)
	s.DrawBullets(win)
	s.player.Draw(win, dt, s.isPaused)
	for _, b := range s.bots {
		b.Draw(win, dt, s.isPaused || s.isTimeStopBonus)
	}
	for _, explosion := range s.explosions {
		explosion.Draw(win, dt, s.isPaused)
	}
	if s.activeBonus != nil {
		s.activeBonus.Draw(win, dt)
	}
	s.rSide.Draw(win)
}

func (s *PlaygroundState) bonusUpdate() {
	if s.activeBonus != nil {
		bonusR := Rect(s.activeBonus.pos, BonusSize, BonusSize)
		playerR := Rect(s.player.pos, TankSize, TankSize)
		if playerR.Intersect(bonusR) != pixel.ZR {
			isLifeBonus := false
			switch s.activeBonus.bonusType {
			case ImmunityBonus:
				s.player.MakeImmune(time.Second * 10)
			case TimeStopBonus:
				s.isTimeStopBonus = true
				s.timeStopDuration = 0
			case HQArmorBonus:
				s.isArmoredHQBonus = true
				s.armoredHQDuration = 0
			case UpgradeBonus:
				s.player.Upgrade()
			case AnnihilationBonus:
				s.annihilateBots()
			case LifeBonus:
				isLifeBonus = true
				if s.player.lives < 9 {
					s.player.lives++
				}
			}
			if isLifeBonus {
				sfx.PlayBonusTakenLife()
			} else {
				sfx.PlayBonusTakenOther()
			}
			s.activeBonus = nil
		}
	}
}

func (s *PlaygroundState) destroyBot(id uuid.UUID) {
	sfx.PlayBotDestroyed()
	s.destroyedBots = append(s.destroyedBots, s.bots[id].botType)
	explosion := explosions.NewExplosion(explosions.TankExplosion, s.bots[id].pos)
	s.explosions = append(s.explosions, explosion)
	delete(s.bots, id)
}

func (s *PlaygroundState) annihilateBots() {
	for id := range s.bots {
		s.destroyBot(id)
	}
	s.newBotDuration = 0
}

func (s *PlaygroundState) Tanks() map[uuid.UUID]Tank {
	tanks := make(map[uuid.UUID]Tank)
	tanks[s.player.id] = s.player
	for id, b := range s.bots {
		tanks[id] = b
	}
	return tanks
}

func (s *PlaygroundState) DrawBullets(win *pixelgl.Window) {
	for _, bullet := range s.bullets {
		m := pixel.IM.Moved(bullet.pos).
			Scaled(bullet.pos, Scale).
			Rotated(bullet.pos, bullet.direction.Angle())
		s.bulletSprite.Draw(win, m)
	}
}

func (s *PlaygroundState) isStageCleared() bool {
	return s.stage.IsPoolEmpty() && len(s.bots) == 0
}
