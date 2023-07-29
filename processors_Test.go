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

func stateTest(state *rwp.HWCState) {
	inImg, err := generateTest(state)
	if err != nil {
		return
	}

	writeImageToHWCGFx(state, inImg)
}

// Converts PNG, JPEG and GIF files into raw panel data
func generateTest(state *rwp.HWCState) (image.Image, error) {
	am := state.Processors.Test

	if am.W == 0 || am.W > 500 || am.H == 0 || am.H > 500 {
		return nil, fmt.Errorf("Invalid dimensions")
	}

	srcImg := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{int(am.W), int(am.H)}})

	sizeString := fmt.Sprintf("%dx%d", am.W, am.H)

	face, err := LoadFontFace("embedded/gfx/fonts/SuperStar.ttf", 16)
	log.Should(err)
	if err == nil {
		dc := gg.NewContextForRGBA(srcImg)
		dc.SetFontFace(face)
		dc.SetColor(color.White)

		strWidth, _ := dc.MeasureString(sizeString)
		xOffset := su.ConstrainValue((int(am.W)-int(strWidth))/2, 0, 1000)
		dc.DrawString(sizeString, float64(xOffset), float64(am.H)/2+4)

		// Actually, I think we should strictly render these lines at a fraction like 0.5 to make them hit a pixel line cleanly, but it works...
		dc.DrawLine(0, 0, float64(am.W), 0)
		dc.DrawLine(0, float64(am.H), float64(am.W), float64(am.H))
		dc.DrawLine(0, 0, 0, float64(am.H))
		dc.DrawLine(float64(am.W), 0, float64(am.W), float64(am.H))
		dc.DrawLine(0, 0, float64(am.W), float64(am.H))
		dc.Stroke()
	}

	return srcImg, nil
}
