package rawpanelproc

import (
	"image"
	"image/color"
	"math"
	"strings"

	su "github.com/SKAARHOJ/ibeam-lib-utils"
)

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

func RangeMap(value int, mapRange []int, RMYAxis bool) int {
	if RMYAxis {
		if len(mapRange) >= 2 {

			descreteStepsY := len(mapRange) - 1 // Divides the Y-axis (output) into sections of equal size
			descreteStepSize := float64(1000) / float64(descreteStepsY)
			if value < mapRange[0] {
				return 0
			}
			if value > mapRange[descreteStepsY] {
				return 1000
			}
			for a := 0; a < descreteStepsY; a++ {
				if value >= mapRange[a] && value <= mapRange[a+1] {
					output := float64(descreteStepSize * float64(a))
					if (mapRange[a+1] - mapRange[a]) == 0 {
						continue // otherwise, we will have division by zero
					}
					pct := float64(value-mapRange[a]) / float64(mapRange[a+1]-mapRange[a])
					outInt := int(math.Round(output + (descreteStepSize * pct)))
					//fmt.Println(value, outInt)
					return outInt
				}
			}
		}
	} else {
		if len(mapRange) >= 2 {
			floatInput := float64(value)
			descreteStepsX := len(mapRange) - 1 // Divides the X-axis (input) into sections
			descreteStepSize := float64(1000) / float64(descreteStepsX)

			for a := 0; a < descreteStepsX; a++ {
				floatA := float64(a)
				if floatInput >= descreteStepSize*floatA && floatInput <= descreteStepSize*(floatA+1) { // Check if input is within range of current step on x-axis
					output := float64(mapRange[a])
					subRange := float64(mapRange[a+1] - mapRange[a])

					pct := (floatInput - descreteStepSize*floatA) / descreteStepSize
					outInt := int(math.Round(output + (subRange * pct)))
					if outInt > 1000 {
						outInt = 1000
					}
					if outInt < 0 {
						outInt = 0
					}

					// Andreas: I have kept this log message here (commented out) as it's extremely usefull when debugging value ranges for Audio meters, between diffrent manufactures.
					// log.Warn(input," ",descreteSteps, " ", descreteStepSize, " ", a, " ", mapRange[a], " ", output, "+(", subRange,"*",pct,")=", outInt)
					return outInt
				}
			}
		}
	}
	return value
}

func GenerateRangeMap(RangeMapping string) []int {
	rangeMap := []int{0, 1000}
	if RangeMapping != "" {
		stringSplit := strings.Split(RangeMapping, ",")
		rangeMap = make([]int, len(stringSplit))
		for i, s := range stringSplit {
			rangeMap[i] = su.Intval(strings.TrimSpace(s))
		}
	}

	return rangeMap
}
