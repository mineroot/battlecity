package main

import (
	"battlecity/game"
	"battlecity/game/explosions"
	"battlecity/game/sfx"
	"bytes"
	"embed"
	"fmt"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"image"
	_ "image/png"
	"math/rand"
	"time"
)

//go:embed assets/stages/*
var stagesConfigs embed.FS

//go:embed assets/sfx/*
var sfxFiles embed.FS

//go:embed assets/spritesheet.png
var spritesheetPng []byte

//go:embed assets/PressStart.ttf
var defaultFontTtf []byte

func loadSpritesheet() (pixel.Picture, error) {
	reader := bytes.NewReader(spritesheetPng)
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

func loadFont() (font.Face, error) {
	f, err := truetype.Parse(defaultFontTtf)
	if err != nil {
		return nil, err
	}

	return truetype.NewFace(f, &truetype.Options{
		Size:              36,
		GlyphCacheEntries: 1,
	}), nil
}

func run() {
	rand.Seed(time.Now().UnixNano())
	cfg := pixelgl.WindowConfig{
		Title:  "Battle City 2022",
		Bounds: pixel.R(0, 0, 1024, 960),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	err = sfx.Init(sfxFiles)
	if err != nil {
		panic(err)
	}
	spritesheet, err := loadSpritesheet()
	if err != nil {
		panic(err)
	}
	defaultFont, err := loadFont()
	if err != nil {
		panic(err)
	}
	explosions.InnitExplosionFrames(spritesheet, game.Scale)
	g := game.NewGame(game.StateConfig{
		Spritesheet:   spritesheet,
		DefaultFont:   defaultFont,
		StagesConfigs: stagesConfigs,
		WindowBounds:  cfg.Bounds,
	})

	secondTick := time.Tick(time.Second)
	frames := 0
	last := time.Now()
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		g.Run(win, dt)
		win.Update()

		frames++
		select {
		case <-secondTick:
			win.SetTitle(fmt.Sprintf("%s | FPS: %d", cfg.Title, frames))
			frames = 0
		default:
		}
	}
}

func main() {
	pixelgl.Run(run)
}
