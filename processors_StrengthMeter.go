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

func stateStrengthMeter(state *rwp.HWCState) {
	inImg, err := generateStrengthMeter(state)
	if err != nil {
		return
	}

	writeImageToHWCGFx(state, inImg)
}

func generateStrengthMeter(state *rwp.HWCState) (image.Image, error) {

	am := state.Processors.StrengthMeter

	if am.W == 0 || am.W > 500 || am.H == 0 || am.H > 500 {
		return nil, fmt.Errorf("Invalid dimensions")
	}

	srcImg := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{int(am.W), int(am.H)}})

	rangeMap := GenerateRangeMap(am.RangeMapping)

	valueStringRelatedOffset := 0
	if am.ValueString != "" {
		valueStringRelatedOffset = 10
		face, err := LoadFontFace("embedded/gfx/fonts/SuperStar.ttf", 16)
		log.Should(err)
		if err == nil {
			dc := gg.NewContextForRGBA(srcImg)
			dc.SetFontFace(face)
			dc.SetColor(color.White)

			dc.RotateAbout(gg.Radians(90), float64(am.W/2), float64(am.W/2))
			strWidth, _ := dc.MeasureString(am.ValueString)
			//fmt.Println(strWidth)
			xOffset := su.ConstrainValue((int(am.H)-int(strWidth))/2+1, 0, 1000)
			dc.DrawString(am.ValueString, float64(xOffset), float64(am.W)-2)
		}
	}

	if am.Title != "" {
		face, err := LoadFontFace("embedded/gfx/fonts/m5x7.ttf", 16)
		log.Should(err)
		if err == nil {
			dc := gg.NewContextForRGBA(srcImg)
			dc.SetFontFace(face)
			dc.SetColor(color.White)
			dc.DrawLine(3+float64(valueStringRelatedOffset), 10.5, float64(am.W)-4, 10.5)
			dc.Stroke()

			strWidth, _ := dc.MeasureString(am.Title)
			xOffset := su.ConstrainValue(valueStringRelatedOffset+(int(am.W)-valueStringRelatedOffset-int(strWidth))/2, 0, 1000)
			dc.DrawString(am.Title, float64(xOffset), 9)
		}
	}

	if am.Data1 >= 0 {
		barW := int(am.W) - 10 - valueStringRelatedOffset
		barH := 5
		wedge := 10
		//fmt.Println(RangeMap(Data1, rangeMap)*barW/1000, 13, 24, barW, barH)
		Strength_bar(srcImg, RangeMap(int(am.Data1), rangeMap, am.RMYAxis)*barW/1000, 5+valueStringRelatedOffset, int(am.H)-barH, barW, barH, wedge)
	}

	return srcImg, nil
}
