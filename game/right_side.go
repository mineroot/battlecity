package game

import (
	"fmt"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font"
	"image/color"
	"math"
)

type RSideData struct {
	stageNum         int
	firstPlayerLives int
	botsPullLen      int
}

type RSide struct {
	batch           *pixel.Batch
	needsRedraw     bool
	firstPlayerIcon *pixel.Sprite
	livesIcon       *pixel.Sprite
	stageIcon       *pixel.Sprite
	botIcon         *pixel.Sprite
	atlas           *text.Atlas
	stageTxt        *text.Text
	livesTxt        *text.Text
	data            *RSideData
}

func NewRightSide(spritesheet pixel.Picture, font font.Face) *RSide {
	r := new(RSide)
	r.batch = pixel.NewBatch(&pixel.TrianglesData{}, spritesheet)
	r.firstPlayerIcon = pixel.NewSprite(spritesheet, pixel.R(376, 112, 392, 120))
	r.livesIcon = pixel.NewSprite(spritesheet, pixel.R(376, 104, 384, 112))
	r.stageIcon = pixel.NewSprite(spritesheet, pixel.R(376, 56, 392, 72))
	r.botIcon = pixel.NewSprite(spritesheet, pixel.R(320, 56, 328, 64))
	r.atlas = text.NewAtlas(font, text.ASCII)
	return r
}

func (r *RSide) Update(data RSideData) {
	if r.data == nil || *r.data != data {
		r.needsRedraw = true
	}
	if r.data == nil || r.data.firstPlayerLives != data.firstPlayerLives {
		r.livesTxt = text.New(pixel.V(
			29*BlockSize*Scale+36,
			10*BlockSize*Scale+30,
		), r.atlas)
		r.livesTxt.Color = colornames.Black
		_, _ = fmt.Fprintln(r.livesTxt, fmt.Sprintf("%d", data.firstPlayerLives))
	}
	if r.data == nil || r.data.stageNum != data.stageNum {
		r.stageTxt = text.New(pixel.V(
			29*BlockSize*Scale+36,
			3*BlockSize*Scale+30,
		), r.atlas)
		r.stageTxt.Color = colornames.Black
		_, _ = fmt.Fprintln(r.stageTxt, fmt.Sprintf("%d", data.stageNum))
	}
	r.data = &data
}

func (r *RSide) Draw(win *pixelgl.Window) {
	r.drawBatch(win)
	r.livesTxt.Draw(win, pixel.IM.Scaled(r.livesTxt.Orig, 0.91))
	r.stageTxt.Draw(win, pixel.IM.Scaled(r.stageTxt.Orig, 0.91))
}

func (r *RSide) drawBatch(win *pixelgl.Window) {
	if !r.needsRedraw {
		r.batch.Draw(win)
		return
	}
	r.batch.Clear()

	//bg
	rightRect := imdraw.New(nil)
	rightRect.Color = color.RGBA{R: 99, G: 99, B: 99, A: 1}
	rightRect.Push(pixel.V(BlockSize*Scale*30, 0), pixel.V(BlockSize*Scale*32, BlockSize*Scale*30))
	rightRect.Rectangle(0)
	rightRect.Draw(r.batch)

	// bots icons
	for i := 0; i < r.data.botsPullLen; i++ {
		yStart := 26*BlockSize*Scale + BlockSize*Scale/2
		row := math.Mod(float64(i), 2)
		botIconPos := pixel.V(
			29*BlockSize*Scale+BlockSize*Scale/2+row*BlockSize*Scale,
			yStart-BlockSize*Scale*float64(i/2),
		)
		r.botIcon.Draw(r.batch, pixel.IM.Moved(botIconPos).Scaled(botIconPos, Scale))
	}

	// first player lives
	firstPlayerIconPos := pixel.V(
		29*BlockSize*Scale+r.firstPlayerIcon.Frame().W()*Scale/2,
		12*BlockSize*Scale+r.firstPlayerIcon.Frame().H()*Scale/2,
	)
	r.firstPlayerIcon.Draw(r.batch, pixel.IM.Moved(firstPlayerIconPos).Scaled(firstPlayerIconPos, Scale))

	// lives icon
	livesIconPos := pixel.V(
		29*BlockSize*Scale+r.livesIcon.Frame().W()*Scale/2,
		11*BlockSize*Scale+r.livesIcon.Frame().H()*Scale/2,
	)
	r.livesIcon.Draw(r.batch, pixel.IM.Moved(livesIconPos).Scaled(livesIconPos, Scale))

	// stage icon
	stageIconPos := pixel.V(
		29*BlockSize*Scale+r.stageIcon.Frame().W()*Scale/2,
		5*BlockSize*Scale+r.stageIcon.Frame().H()*Scale/2,
	)
	r.stageIcon.Draw(r.batch, pixel.IM.Moved(stageIconPos).Scaled(stageIconPos, Scale))

	r.batch.Draw(win)
	r.needsRedraw = false
}
