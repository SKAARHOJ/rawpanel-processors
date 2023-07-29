package rawpanelproc

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"

	su "github.com/SKAARHOJ/ibeam-lib-utils"
	rwp "github.com/SKAARHOJ/rawpanel-lib/ibeam_rawpanel"
	"github.com/fogleman/gg"
	log "github.com/s00500/env_logger"
)

func stateAudioMeter(state *rwp.HWCState) {
	inImg, err := generateAudioMeter(state)
	if err != nil {
		return
	}

	writeImageToHWCGFx(state, inImg)
}

func generateAudioMeter(state *rwp.HWCState) (image.Image, error) {

	var srcImg image.Image
	var err error
	am := state.Processors.AudioMeter

	// Perfect linear range for this type of VU meter is
	// 0,77,154,231,308,385,462,538,615,692,769,846,923,1000
	// to match
	// -60, -54, -48, -42, -36, -30, -24, -18, -12, -6,   0,  6, 12
	// -90, -60, -54, -48, -42, -36, -30, -24, -18, -12, -6, -3,  0
	rangeMap := GenerateRangeMap(am.RangeMapping)

	switch am.MeterType {
	case rwp.ProcAudioMeter_Fixed176x32, rwp.ProcAudioMeter_Fixed176x32w:

		fileName := "embedded/gfx/widgets/VUMeter_12dB_176x32.png" // Default fallback image
		if am.MeterType == rwp.ProcAudioMeter_Fixed176x32w {
			fileName = "embedded/gfx/widgets/VUMeter_wide_12dB_176x32.png"
		}

		if ResourceFileExist(fileName) {
			imageBytes := ReadResourceFile(fileName)

			myReader := bytes.NewReader(imageBytes)
			srcImg, err = png.Decode(myReader)
			if err == nil {
				if am.Title != "" {
					face, err := LoadFontFace("embedded/gfx/fonts/m5x7.ttf", 16)
					log.Should(err)
					if err == nil {
						dc := gg.NewContextForRGBA(srcImg.(*image.RGBA))
						dc.SetFontFace(face)
						dc.SetColor(color.White)

						dc.RotateAbout(gg.Radians(90), 176/2, 176/2)
						strWidth, _ := dc.MeasureString(am.Title)
						xOffset := su.ConstrainValue((32-int(strWidth))/2, 0, 16)
						dc.DrawString(am.Title, float64(xOffset), 9)
					}
				}

				switch am.MeterType {
				case rwp.ProcAudioMeter_Fixed176x32w:
					if am.Data1 >= 0 {
						Fixed176x32_blocks(srcImg, RangeMap(int(am.Data1), rangeMap, am.RMYAxis)*144/1000, 14, 14, am.Mono)
					}
					if am.Data2 >= 0 && !am.Mono {
						Fixed176x32_blocks(srcImg, RangeMap(int(am.Data2), rangeMap, am.RMYAxis)*144/1000, 14, 14+10, false)
					}
					if am.Peak1 >= 0 {
						Fixed176x32_peakblock(srcImg, RangeMap(int(am.Peak1), rangeMap, am.RMYAxis)*144/1000, 14, 14, am.Mono)
					}
					if am.Peak2 >= 0 && !am.Mono {
						Fixed176x32_peakblock(srcImg, RangeMap(int(am.Peak2), rangeMap, am.RMYAxis)*144/1000, 14, 14+10, false)
					}
				default:
					if am.Data1 >= 0 {
						Fixed176x32_bar(srcImg, RangeMap(int(am.Data1), rangeMap, am.RMYAxis)*144/1000, 13, 24, am.Mono)
					}
					if am.Data2 >= 0 && !am.Mono {
						Fixed176x32_bar(srcImg, RangeMap(int(am.Data2), rangeMap, am.RMYAxis)*144/1000, 13, 24+5, false)
					}
					if am.Peak1 >= 0 {
						Fixed176x32_peak(srcImg, RangeMap(int(am.Peak1), rangeMap, am.RMYAxis)*144/1000, 13, 24, am.Mono)
					}
					if am.Peak2 >= 0 && !am.Mono {
						Fixed176x32_peak(srcImg, RangeMap(int(am.Peak2), rangeMap, am.RMYAxis)*144/1000, 13, 24+5, false)
					}
				}
				return srcImg, nil
			}
		} else {
			log.Println("nou foudn")
		}
	default:
		if am.W == 0 || am.W > 500 || am.H == 0 || am.H > 500 {
			return nil, fmt.Errorf("Invalid dimensions")
		}

		srcImg := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{int(am.W), int(am.H)}})
		yOff := 0

		if am.Title != "" {
			face, err := LoadFontFace("embedded/gfx/fonts/m5x7.ttf", 16)
			log.Should(err)
			if err == nil {
				dc := gg.NewContextForRGBA(srcImg)
				dc.SetFontFace(face)
				dc.SetColor(color.White)
				dc.DrawLine(3, 10.5, float64(am.W)-4, 10.5)
				dc.Stroke()

				strWidth, _ := dc.MeasureString(am.Title)
				xOffset := su.ConstrainValue((int(am.W)-int(strWidth))/2, 0, 1000)
				dc.DrawString(am.Title, float64(xOffset), 9)

				yOff += 11
			}
		}

		barAreaHeight := int(am.H) - yOff
		barDist := barAreaHeight / 3
		barHeight := barAreaHeight - barDist
		if !am.Mono {
			barAreaHeight = barAreaHeight / 2

			barDist = barAreaHeight / 3
			barHeight = barAreaHeight - barDist
		}

		if am.Data1 >= 0 {
			VU_blocks(srcImg, RangeMap(int(am.Data1), rangeMap, am.RMYAxis)*int(am.W)/1000, 0, int(am.W), yOff+barDist, barHeight)
		}
		if am.Data2 >= 0 && !am.Mono {
			VU_blocks(srcImg, RangeMap(int(am.Data2), rangeMap, am.RMYAxis)*int(am.W)/1000, 0, int(am.W), yOff+barDist*2+barHeight, barHeight)
		}
		if am.Peak1 >= 0 {
			VU_peakblock(srcImg, RangeMap(int(am.Peak1), rangeMap, am.RMYAxis)*int(am.W)/1000, 0, int(am.W), yOff+barDist, barHeight)
		}
		if am.Peak2 >= 0 && !am.Mono {
			VU_peakblock(srcImg, RangeMap(int(am.Peak2), rangeMap, am.RMYAxis)*int(am.W)/1000, 0, int(am.W), yOff+barDist*2+barHeight, barHeight)
		}

		return srcImg, nil
	}

	return nil, fmt.Errorf("Invalid meter type")
}
