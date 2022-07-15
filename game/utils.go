package game

import "github.com/faiface/pixel"

func Rect(pos pixel.Vec, w float64, h float64) pixel.Rect {
	w, h = w*Scale/2, h*Scale/2
	return pixel.R(pos.X-w, pos.Y-h, pos.X+w, pos.Y+h)
}

func MRound(rounder func(float64) float64, n float64, multiple float64) float64 {
	return multiple * rounder(n/multiple)
}
