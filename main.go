package main

import (
	"battlecity/game"
	"embed"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

//go:embed assets/stages/*
var stagesConfigs embed.FS

func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

func run() {
	//monitor := pixelgl.PrimaryMonitor()
	cfg := pixelgl.WindowConfig{
		Title:  "Battle City 2022",
		Bounds: pixel.R(0, 0, 1024, 960),
		VSync:  true,
		//Monitor: monitor,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	spritesheet, err := loadPicture("assets/spritesheet.png")
	if err != nil {
		panic(err)
	}
	ml := game.CreateMainLoop(&spritesheet, stagesConfigs)

	secondTick := time.Tick(time.Second)
	frames := 0
	last := time.Now()
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		ml.Run(win, dt)

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
