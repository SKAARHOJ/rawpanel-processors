package rawpanelproc

import (
	"fmt"
	"image"
	"image/color"

	su "github.com/SKAARHOJ/ibeam-lib-utils"
	rwp "github.com/SKAARHOJ/rawpanel-lib/ibeam_rawpanel"
	"github.com/fogleman/gg"
	log "github.com/s00500/env_logger"
)

func stateUniText(state *rwp.HWCState) {
	inImg, err := generateUniText(state)
	if err != nil {
		return
	}

	writeImageToHWCGFx(state, inImg)
}

// Converts PNG, JPEG and GIF files into raw panel data
func generateUniText(state *rwp.HWCState) (image.Image, error) {
	am := state.Processors.UniText

	if am.W == 0 || am.W > 500 || am.H == 0 || am.H > 500 {
		return nil, fmt.Errorf("Invalid dimensions")
	}

	srcImg := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{int(am.W), int(am.H)}})

	fontSize := float64(16) // Cannot change, because we use Unifont.ttf which is fixed size 16x16
	face, err := LoadFontFace("embedded/gfx/fonts/Unifont.ttf", fontSize)
	log.Should(err)
	if err == nil {

		dc := gg.NewContextForRGBA(srcImg)
		dc.SetFontFace(face)

		// Title:
		titleComp := 0
		if am.Title != "" {
			dc.SetColor(color.White)

			if am.SolidHeaderBar {
				dc.DrawRectangle(0, 0, float64(am.W), fontSize)
				dc.Fill()

				srcImg.Set(0, 0, color.RGBA{0, 0, 0, 255})
				srcImg.Set(int(am.W)-1, 0, color.RGBA{0, 0, 0, 255})
				srcImg.Set(0, 15, color.RGBA{0, 0, 0, 255})
				srcImg.Set(int(am.W)-1, 15, color.RGBA{0, 0, 0, 255})

				dc.SetColor(color.Black)
			} else {
				dc.DrawLine(2, 15.5, float64(am.W)-2, 15.5)
				dc.Stroke()
			}

			strWidth, _ := dc.MeasureString(am.Title)
			xOffset := su.ConstrainValue((int(am.W)-int(strWidth))/2, 0, 1000)
			dc.DrawString(am.Title, float64(xOffset), 14)
			titleComp = int(fontSize)
		}

		dc.SetColor(color.White)

		// Add all possible lines of text and remove empty lines at the end
		lines := []string{am.Textline1, am.Textline2, am.Textline3, am.Textline4}
		for i := len(lines) - 1; i >= 0; i-- {
			if lines[i] == "" {
				lines = lines[:i]
			} else {
				break
			}
		}
		numberOfLines := len(lines)
		if numberOfLines > 0 {
			lineComp := su.ConstrainValue((int(am.H)-titleComp-numberOfLines*int(fontSize))/(numberOfLines*2), 0, 100)

			for i, line := range lines {
				if line == "" {
					continue
				}
				strWidth, _ := dc.MeasureString(line)
				xOffset := getAlignedXOffset(am, strWidth)
				dc.DrawString(line, float64(xOffset), float64(titleComp+14+lineComp*(i*2+1)+int(fontSize)*i))
			}
		}
	}

	return srcImg, nil
}

func getAlignedXOffset(am *rwp.ProcUniText, textWidth float64) int {
	switch am.Align {
	case rwp.ProcUniText_LEFT:
		return 0
	case rwp.ProcUniText_RIGHT:
		return su.ConstrainValue(int(am.W)-int(textWidth), -1000, 1000) // Accept cropping, but not too far right
	case rwp.ProcUniText_CENTER:
		fallthrough
	default:
		return su.ConstrainValue((int(am.W)-int(textWidth))/2, 0, 1000) // Center, but not too far left
	}
}
