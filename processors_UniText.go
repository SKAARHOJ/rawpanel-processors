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

	face, err := LoadFontFace("embedded/gfx/fonts/Unifont.ttf", 16)
	log.Should(err)
	if err == nil {

		dc := gg.NewContextForRGBA(srcImg)
		dc.SetFontFace(face)

		// Title:
		titleComp := 0
		if am.Title != "" {
			dc.SetColor(color.White)

			if am.SolidHeaderBar {
				dc.DrawRectangle(0, 0, float64(am.W), 16)
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
			dc.DrawString(am.Title, float64(xOffset), 14) //float64(am.H)/2+6/2
			titleComp = 16
		}

		dc.SetColor(color.White)
		line2Comp := su.Qint(am.Textline2 != "", 8+su.ConstrainValue((int(am.H)-titleComp-32)/4, 0, 100), 0)
		if am.Textline1 != "" {
			strWidth, _ := dc.MeasureString(am.Textline1)
			xOffset := su.ConstrainValue((int(am.W)-int(strWidth))/2, 0, 1000)
			dc.DrawString(am.Textline1, float64(xOffset), float64((int(am.H)-titleComp)/2+6/2+3+titleComp-line2Comp))
		}
		if am.Textline2 != "" {
			strWidth, _ := dc.MeasureString(am.Textline2)
			xOffset := su.ConstrainValue((int(am.W)-int(strWidth))/2, 0, 1000)
			dc.DrawString(am.Textline2, float64(xOffset), float64((int(am.H)-titleComp)/2+6/2+3+titleComp+line2Comp))
		}

		// Actually, I think we should strictly render these lines at a fraction like 0.5 to make them hit a pixel line cleanly, but it works...
	}

	return srcImg, nil
}
