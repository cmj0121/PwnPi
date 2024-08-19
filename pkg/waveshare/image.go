// The image utility functions.
package waveshare

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	_ "image/png"

	"github.com/disintegration/imaging"
	"github.com/rs/zerolog/log"
)

// The list of available images.
const (
	IMG_RPI_ICON = "assets/images/raspberry-pi.png"
)

// Get the raw image data from the embedded filesystem.
func (w *WaveShare) image(path string) (image.Image, error) {
	data, err := w.images.ReadFile(path)
	if err != nil {
		log.Warn().Str("path", path).Err(err).Msg("failed to read the image file")
		return nil, err
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		log.Warn().Str("path", path).Err(err).Msg("failed to decode the image file")
		return nil, err
	}

	// remove the transparency from the image and set background color to white.
	bg := image.NewRGBA(img.Bounds())
	draw.Draw(bg, bg.Bounds(), image.White, image.ZP, draw.Src)
	img = imaging.Overlay(bg, img, image.Pt(0, 0), 1.0)

	return img, nil
}

// Resize the image to the specified width and height and keep the aspect ratio.
func (w *WaveShare) resizeImage(img image.Image, width, height, padding int) image.Image {
	r1 := float64(width) / float64(height)
	r2 := float64(img.Bounds().Dx()) / float64(img.Bounds().Dy())

	switch {
	case r2 > r1:
		img = imaging.Resize(img, width-padding*2, 0, imaging.Lanczos)
	case r2 < r1:
		img = imaging.Resize(img, 0, height-padding*2, imaging.Lanczos)
	}

	// move the image to the center of the display.
	bg := imaging.New(TP2in13_WIDTH, TP2in13_HEIGHT, color.White)
	draw.Draw(bg, bg.Bounds(), image.White, image.ZP, draw.Src)
	img = imaging.PasteCenter(bg, img)

	return img
}

// Show the image on the WaveShare E-Ink display.
func (w *WaveShare) showImage(path string, padding int) error {
	img, err := w.image(path)
	if err != nil {
		log.Warn().Str("path", path).Err(err).Msg("failed to get the image")
		return err
	}

	// rotate and resize the passed image.
	img = imaging.Rotate270(img)
	img = w.resizeImage(img, TP2in13_WIDTH, TP2in13_HEIGHT, padding)

	// convert the image to the 1-bit color image.
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	wire := (width + 7) / 8

	// remove the transparency from the image and make it as 1-bit color image.
	pixels := make([]byte, wire*height)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		byteIdx, bitIdx := 0, 7
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray := color.GrayModel.Convert(img.At(x, y)).(color.Gray).Y

			var bit byte
			switch {
			case gray < 128:
				bit = 0
			default:
				bit = 1
			}

			pixels[byteIdx+y*wire] |= bit << bitIdx
			bitIdx--
			if bitIdx < 0 {
				byteIdx++
				bitIdx = 7
			}
		}
	}

	return w.showPixel(true, pixels...)
}
