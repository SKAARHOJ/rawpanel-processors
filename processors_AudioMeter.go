package rawpanelproc

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"

	su "github.com/SKAARHOJ/ibeam-lib-utils"
	rwp "github.com/SKAARHOJ/rawpanel-lib/ibeam_rawpanel"
	"github.com/fogleman/gg"
	log "github.com/s00500/env_logger"
)

// Converts PNG, JPEG and GIF files into raw panel data
func stateAudioMeter(state *rwp.HWCState) {

	var srcImg image.Image
	var err error
	am := state.Processors.AudioMeter

	Mono := am.Mono
	var Data1, Data2, Data3, Data4 int

	// Perfect linear range for this type of VU meter is
	// 0,77,154,231,308,385,462,538,615,692,769,846,923,1000
	// to match
	// -60, -54, -48, -42, -36, -30, -24, -18, -12, -6,   0,  6, 12
	// -90, -60, -54, -48, -42, -36, -30, -24, -18, -12, -6, -3,  0
	//rangeMap := widget.getRangeMap()
	rangeMap := []int{0, 1000}

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
					if Data1 >= 0 {
						Fixed176x32_blocks(srcImg, RangeMap(Data1, rangeMap)*144/1000, 14, 14, Mono)
					}
					if Data2 >= 0 && !Mono {
						Fixed176x32_blocks(srcImg, RangeMap(Data2, rangeMap)*144/1000, 14, 14+10, false)
					}
					if Data3 >= 0 {
						Fixed176x32_peakblock(srcImg, RangeMap(Data3, rangeMap)*144/1000, 14, 14, Mono)
					}
					if Data4 >= 0 && !Mono {
						Fixed176x32_peakblock(srcImg, RangeMap(Data4, rangeMap)*144/1000, 14, 14+10, false)
					}
				default:
					if Data1 >= 0 {
						Fixed176x32_bar(srcImg, RangeMap(Data1, rangeMap)*144/1000, 13, 24, Mono)
					}
					if Data2 >= 0 && !Mono {
						Fixed176x32_bar(srcImg, RangeMap(Data2, rangeMap)*144/1000, 13, 24+5, false)
					}
					if Data3 >= 0 {
						Fixed176x32_peak(srcImg, RangeMap(Data3, rangeMap)*144/1000, 13, 24, Mono)
					}
					if Data4 >= 0 && !Mono {
						Fixed176x32_peak(srcImg, RangeMap(Data4, rangeMap)*144/1000, 13, 24+5, false)
					}
				}
				//	return srcImg, nil
			}
		}
	default:
		if am.W == 0 || am.W > 500 || am.H == 0 || am.H > 500 {
			return
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
		if !Mono {
			barAreaHeight = barAreaHeight / 2

			barDist = barAreaHeight / 3
			barHeight = barAreaHeight - barDist
		}

		if Data1 >= 0 {
			VU_blocks(srcImg, RangeMap(Data1, rangeMap)*int(am.W)/1000, 0, int(am.W), yOff+barDist, barHeight)
		}
		if Data2 >= 0 && !Mono {
			VU_blocks(srcImg, RangeMap(Data2, rangeMap)*int(am.W)/1000, 0, int(am.W), yOff+barDist*2+barHeight, barHeight)
		}
		if Data3 >= 0 {
			VU_peakblock(srcImg, RangeMap(Data3, rangeMap)*int(am.W)/1000, 0, int(am.W), yOff+barDist, barHeight)
		}
		if Data4 >= 0 && !Mono {
			VU_peakblock(srcImg, RangeMap(Data4, rangeMap)*int(am.W)/1000, 0, int(am.W), yOff+barDist*2+barHeight, barHeight)
		}

		// Set the new image:
		//state.HWCGfx = &img
		state.HWCText = nil
		state.Processors.AudioMeter = nil

		//return srcImg, nil
	}
}

func Fixed176x32_bar(srcImg image.Image, barLength int, Xoffset int, Yoffset int, mono bool) {
	if barLength >= 1 {
		width := 3
		if mono {
			width = 8
		}
		if barLength > 144 {
			barLength = 144 // limit...
		}
		col := color.RGBA{255, 255, 255, 255}
		blackOut := color.RGBA{0, 0, 0, 255}
		for a := 0; a < width; a++ {
			HLine(srcImg.(*image.RGBA), Xoffset, Yoffset+a, Xoffset+barLength-1, col)
		}
		if barLength%12 == 0 {
			HLine(srcImg.(*image.RGBA), Xoffset+barLength-1, Yoffset-1, Xoffset+barLength-1, blackOut)
			HLine(srcImg.(*image.RGBA), Xoffset+barLength-1, Yoffset+3, Xoffset+barLength-1, blackOut)
		}
	}
}
func Fixed176x32_peak(srcImg image.Image, peakPos int, Xoffset int, Yoffset int, mono bool) {
	if peakPos >= 1 {
		width := 3
		if mono {
			width = 8
		}
		if peakPos > 144 {
			peakPos = 144 // limit...
		}
		col := color.RGBA{255, 255, 255, 255}
		blackOut := color.RGBA{0, 0, 0, 255}
		for a := 0; a < width; a++ {
			HLine(srcImg.(*image.RGBA), Xoffset+peakPos-1, Yoffset+a, Xoffset+peakPos-1, col)
		}
		if peakPos%12 == 0 {
			HLine(srcImg.(*image.RGBA), Xoffset+peakPos-1, Yoffset-1, Xoffset+peakPos-1, blackOut)
			HLine(srcImg.(*image.RGBA), Xoffset+peakPos-1, Yoffset+3, Xoffset+peakPos-1, blackOut)
		}
	}
}
func Fixed176x32_blocks(srcImg image.Image, barLength int, Xoffset int, Yoffset int, mono bool) {
	if barLength >= 1 {
		width := 8
		if mono {
			width = 18
		}
		if barLength > 144 {
			barLength = 144 // limit...
		}
		blocks := (barLength + 3) >> 2 // Ceil-Divide by 4
		col := color.RGBA{255, 255, 255, 255}
		for a := 0; a < width; a++ {
			for b := 0; b < blocks; b++ {
				HLine(srcImg.(*image.RGBA), Xoffset+(b<<2), Yoffset+a, Xoffset+((b+1)<<2)-2, col)
			}
		}
	}
}
func Fixed176x32_peakblock(srcImg image.Image, peakPos int, Xoffset int, Yoffset int, mono bool) {
	if peakPos >= 1 {
		width := 8
		if mono {
			width = 18
		}
		if peakPos > 144 {
			peakPos = 144 // limit...
		}
		blocks := (peakPos + 3) >> 2 // Ceil-Divide by 4
		col := color.RGBA{255, 255, 255, 255}
		for a := 0; a < width; a++ {
			HLine(srcImg.(*image.RGBA), Xoffset+((blocks-1)<<2), Yoffset+a, Xoffset+((blocks)<<2)-2, col)
		}
	}
}
func VU_blocks(srcImg image.Image, barLength int, Xoffset int, Xwidth int, Yoffset int, BarHeight int) {
	if barLength >= 1 {
		if barLength > Xwidth {
			barLength = Xwidth // limit...
		}
		blocks := (barLength + 3) >> 2 // Ceil-Divide by 4
		col := color.RGBA{255, 255, 255, 255}
		for a := 0; a < BarHeight; a++ {
			for b := 0; b < blocks; b++ {
				HLine(srcImg.(*image.RGBA), Xoffset+(b<<2), Yoffset+a, Xoffset+((b+1)<<2)-2, col)
			}
		}
	}
}
func VU_peakblock(srcImg image.Image, peakPos int, Xoffset int, Xwidth int, Yoffset int, BarHeight int) {
	if peakPos >= 1 {
		if peakPos > Xwidth {
			peakPos = Xwidth // limit...
		}
		blocks := (peakPos + 3) >> 2 // Ceil-Divide by 4
		col := color.RGBA{255, 255, 255, 255}
		for a := 0; a < BarHeight; a++ {
			HLine(srcImg.(*image.RGBA), Xoffset+((blocks-1)<<2), Yoffset+a, Xoffset+((blocks)<<2)-2, col)
		}
	}
}
func Strength_bar(srcImg image.Image, barLength int, Xoffset int, Yoffset int, barW int, barH int, wedge int) {
	if barLength >= 1 {
		if barLength > barW {
			barLength = barW // limit...
		}

		for a := 0; a < barLength; a++ {
			if a%3 > 0 {
				gray := uint8(su.ConstrainValue(a*255/barLength, 128, 255))
				col := color.RGBA{gray, gray, gray, 255}
				wedgingFactor := (wedge * a / barW)
				VLine(srcImg.(*image.RGBA), Xoffset+a, Yoffset-wedgingFactor, Yoffset+barH, col)
			}
		}
	}
}

// HLine draws a horizontal line
func HLine(img *image.RGBA, x1, y, x2 int, col color.RGBA) {
	for ; x1 <= x2; x1++ {
		img.Set(x1, y, col)
	}
}

// VLine draws a veritcal line
func VLine(img *image.RGBA, x, y1, y2 int, col color.RGBA) {
	for ; y1 <= y2; y1++ {
		img.Set(x, y1, col)
	}
}

func RangeMap(input int, mapRange []int) int {
	if len(mapRange) >= 2 {
		floatInput := float64(input)
		descreteSteps := len(mapRange) - 1
		descreteStepSize := float64(1000) / float64(descreteSteps)

		for a := 0; a < descreteSteps; a++ {
			floatA := float64(a)
			if floatInput >= descreteStepSize*floatA && floatInput <= descreteStepSize*(floatA+1) { // Check if input is within range
				output := float64(mapRange[a])
				subRange := float64(mapRange[a+1] - mapRange[a])

				pct := (floatInput - descreteStepSize*floatA) / descreteStepSize
				outInt := int(math.Round(output + (subRange * pct)))

				// Andreas: I have kept this log message here (commented out) as it's extremely usefull when debugging value ranges for Audio meters, between diffrent manufactures.
				// log.Warn(input," ",descreteSteps, " ", descreteStepSize, " ", a, " ", mapRange[a], " ", output, "+(", subRange,"*",pct,")=", outInt)
				return outInt
			}
		}
	}

	return input
}
