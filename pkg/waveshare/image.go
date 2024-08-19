// The image utility functions.
package waveshare

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	_ "image/png"
	"io/fs"
	"os"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/rs/zerolog/log"
)

// The list of available images.
const (
	IMG_RPI_ICON = "assets/images/raspberry-pi.png"
)

// The list of available fonts
const (
	FONT_SPACE_MONO_REGULAR     = "assets/fonts/SpaceMono-Regular.ttf"
	FONT_SPACE_MONO_BOLD        = "assets/fonts/SpaceMono-Bold.ttf"
	FONT_SPACE_MONO_ITALIC      = "assets/fonts/SpaceMono-Italic.ttf"
	FONT_SPACE_MONO_BOLD_ITALIC = "assets/fonts/SpaceMono-BoldItalic.ttf"
)

// Get the raw image data from the embedded filesystem.
func (w *WaveShare) image(path string) (image.Image, error) {
	data, err := w.fs.ReadFile(path)
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

func (w *WaveShare) showAssets(path string, padding int) error {
	img, err := w.image(path)
	if err != nil {
		log.Warn().Str("path", path).Err(err).Msg("failed to get the image")
		return err
	}

	return w.showImage(img, padding)
}

// Show the image on the WaveShare E-Ink display.
func (w *WaveShare) showImage(img image.Image, padding int) error {
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

func (w *WaveShare) showText(text string, width, height, fontSize float64) error {
	img, err := w.text2image(text, FONT_SPACE_MONO_BOLD, width, height, fontSize)
	if err != nil {
		log.Warn().Str("text", text).Err(err).Msg("failed to convert the text to image")
		return err
	}

	return w.showImage(img, 8)
}

// Save the text as an image file.
func (w *WaveShare) text2image(text string, font string, width, height, fontSize float64) (image.Image, error) {
	dc := gg.NewContext(int(width), int(height))
	dc.SetRGB(1, 1, 1)
	dc.Clear()

	bytes, err := fs.ReadFile(w.fs, font)
	if err != nil {
		log.Info().Str("font", font).Err(err).Msg("failed to read the font file")
		return nil, err
	}

	// save the font file into the temporary file.
	tmpFont, err := os.CreateTemp(os.TempDir(), "font-*.ttf")
	if err != nil {
		log.Info().Err(err).Msg("failed to create the temporary file")
		return nil, err
	}

	if _, err := tmpFont.Write(bytes); err != nil {
		log.Info().Err(err).Msg("failed to write the font file")
		return nil, err
	}

	// set the font and size.
	if err := dc.LoadFontFace(tmpFont.Name(), float64(fontSize)); err != nil {
		log.Warn().Str("font", font).Err(err).Msg("failed to load the font face")
		return nil, err
	}

	x, y := dc.MeasureString(text)
	x, y = (width-x)/2, (height+y)/2

	// draw the text on the image.
	dc.SetRGB(0, 0, 0)
	dc.DrawString(text, x, y)

	img := dc.Image()
	return img, nil
}
