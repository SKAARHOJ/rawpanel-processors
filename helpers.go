package rawpanelproc

import (
	"image"
	"image/color"
	"math"
	"sort"
	"strings"

	su "github.com/SKAARHOJ/ibeam-lib-utils"
)

// fixed176x32BarSpan is the width in pixels the Fixed176x32 artwork leaves for the bar
// itself. It is both the clamp on how far a bar may be drawn and the span its colours
// are resolved across, so a zone edge lands on the dB tick it was configured for.
const fixed176x32BarSpan = 144

// The meter primitives come in pairs. The original names keep their original signatures
// and paint the plain white these meters have always been drawn in; each *Colored
// variant takes a BarColorizer as well, where nil is that same white. See BarColorizer
// below.

func Fixed176x32_bar(srcImg image.Image, barLength int, Xoffset int, Yoffset int, mono bool) {
	Fixed176x32_barColored(srcImg, barLength, Xoffset, Yoffset, mono, nil)
}
func Fixed176x32_peak(srcImg image.Image, peakPos int, Xoffset int, Yoffset int, mono bool) {
	Fixed176x32_peakColored(srcImg, peakPos, Xoffset, Yoffset, mono, nil)
}
func Fixed176x32_blocks(srcImg image.Image, barLength int, Xoffset int, Yoffset int, mono bool) {
	Fixed176x32_blocksColored(srcImg, barLength, Xoffset, Yoffset, mono, nil)
}
func Fixed176x32_peakblock(srcImg image.Image, peakPos int, Xoffset int, Yoffset int, mono bool) {
	Fixed176x32_peakblockColored(srcImg, peakPos, Xoffset, Yoffset, mono, nil)
}
func VU_blocks(srcImg image.Image, barLength int, Xoffset int, Xwidth int, Yoffset int, BarHeight int) {
	VU_blocksColored(srcImg, barLength, Xoffset, Xwidth, Yoffset, BarHeight, nil)
}
func VU_peakblock(srcImg image.Image, peakPos int, Xoffset int, Xwidth int, Yoffset int, BarHeight int) {
	VU_peakblockColored(srcImg, peakPos, Xoffset, Xwidth, Yoffset, BarHeight, nil)
}

func Fixed176x32_barColored(srcImg image.Image, barLength int, Xoffset int, Yoffset int, mono bool, colorize BarColorizer) {
	if barLength >= 1 {
		width := 3
		if mono {
			width = 8
		}
		if barLength > fixed176x32BarSpan {
			barLength = fixed176x32BarSpan // limit...
		}
		blackOut := color.RGBA{0, 0, 0, 255}
		for a := 0; a < width; a++ {
			HLineColored(srcImg.(*image.RGBA), Xoffset, Yoffset+a, Xoffset+barLength-1, colorize, Xoffset, fixed176x32BarSpan)
		}
		if barLength%12 == 0 {
			HLine(srcImg.(*image.RGBA), Xoffset+barLength-1, Yoffset-1, Xoffset+barLength-1, blackOut)
			HLine(srcImg.(*image.RGBA), Xoffset+barLength-1, Yoffset+3, Xoffset+barLength-1, blackOut)
		}
	}
}
func Fixed176x32_peakColored(srcImg image.Image, peakPos int, Xoffset int, Yoffset int, mono bool, colorize BarColorizer) {
	if peakPos >= 1 {
		width := 3
		if mono {
			width = 8
		}
		if peakPos > fixed176x32BarSpan {
			peakPos = fixed176x32BarSpan // limit...
		}
		// A peak is a single mark rather than a run, so it takes the one colour of the
		// level it marks.
		col := colorAt(colorize, peakPos-1, fixed176x32BarSpan)
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
func Fixed176x32_blocksColored(srcImg image.Image, barLength int, Xoffset int, Yoffset int, mono bool, colorize BarColorizer) {
	if barLength >= 1 {
		width := 8
		if mono {
			width = 18
		}
		if barLength > fixed176x32BarSpan {
			barLength = fixed176x32BarSpan // limit...
		}
		blocks := (barLength + 3) >> 2 // Ceil-Divide by 4
		for a := 0; a < width; a++ {
			for b := 0; b < blocks; b++ {
				// A block is one unit on the meter, so it takes a single colour - from
				// its own left edge - rather than being split across a zone boundary.
				col := colorAt(colorize, b<<2, fixed176x32BarSpan)
				HLine(srcImg.(*image.RGBA), Xoffset+(b<<2), Yoffset+a, Xoffset+((b+1)<<2)-2, col)
			}
		}
	}
}
func Fixed176x32_peakblockColored(srcImg image.Image, peakPos int, Xoffset int, Yoffset int, mono bool, colorize BarColorizer) {
	if peakPos >= 1 {
		width := 8
		if mono {
			width = 18
		}
		if peakPos > fixed176x32BarSpan {
			peakPos = fixed176x32BarSpan // limit...
		}
		blocks := (peakPos + 3) >> 2 // Ceil-Divide by 4
		col := colorAt(colorize, (blocks-1)<<2, fixed176x32BarSpan)
		for a := 0; a < width; a++ {
			HLine(srcImg.(*image.RGBA), Xoffset+((blocks-1)<<2), Yoffset+a, Xoffset+((blocks)<<2)-2, col)
		}
	}
}
func VU_blocksColored(srcImg image.Image, barLength int, Xoffset int, Xwidth int, Yoffset int, BarHeight int, colorize BarColorizer) {
	if barLength >= 1 {
		if barLength > Xwidth {
			barLength = Xwidth // limit...
		}
		blocks := (barLength + 3) >> 2 // Ceil-Divide by 4
		for a := 0; a < BarHeight; a++ {
			for b := 0; b < blocks; b++ {
				col := colorAt(colorize, b<<2, Xwidth)
				HLine(srcImg.(*image.RGBA), Xoffset+(b<<2), Yoffset+a, Xoffset+((b+1)<<2)-2, col)
			}
		}
	}
}
func VU_peakblockColored(srcImg image.Image, peakPos int, Xoffset int, Xwidth int, Yoffset int, BarHeight int, colorize BarColorizer) {
	if peakPos >= 1 {
		if peakPos > Xwidth {
			peakPos = Xwidth // limit...
		}
		blocks := (peakPos + 3) >> 2 // Ceil-Divide by 4
		col := colorAt(colorize, (blocks-1)<<2, Xwidth)
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

// HLineColored draws a horizontal line belonging to a meter bar, taking each pixel's
// colour from where it sits along that bar. barX0 is the bar's own left edge and span
// its full width in pixels, which together turn a pixel into the 0..1000 position a
// BarColorizer expects.
func HLineColored(img *image.RGBA, x1, y, x2 int, colorize BarColorizer, barX0, span int) {
	for ; x1 <= x2; x1++ {
		img.Set(x1, y, colorAt(colorize, x1-barX0, span))
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

// meterWhite is the colour every meter bar was painted before colourisers existed, and
// what a nil BarColorizer still paints.
var meterWhite = color.RGBA{255, 255, 255, 255}

// BarColorizer returns the colour a meter bar or peak indicator is painted with at a
// position along its span.
//
// pos is 0..1000 of the bar's full span rather than a pixel offset. That is deliberate:
// the Fixed176x32 meters draw a 144px bar while the dynamic ones draw one as wide as
// the display, so pixel positions would need different thresholds per meter type and
// per display size. In the 0..1000 domain - the same domain RangeMap outputs - one set
// of stops describes the same zones everywhere, and they line up with the dB scale
// printed on the fixed artwork.
type BarColorizer func(pos int) color.RGBA

// ColorStop is one boundary in a zone colouriser: Col applies from Pos onward, until
// the next stop's Pos. Pos is in the same 0..1000 span domain as BarColorizer's argument.
type ColorStop struct {
	Pos int
	Col color.RGBA
}

// NewZoneColorizer builds a colouriser painting hard-edged zones - the classic
// green/orange/red audio meter - from a set of stops.
//
// Each stop's colour holds until the next one begins, with no interpolation between
// them, so a zone boundary lands exactly on the tick it was configured for. Positions
// below the first stop stay meterWhite: stops recolour a bar from their position
// onward rather than replacing it wholesale, so a lone "red from 920" leaves the rest
// of the meter as it always looked.
//
// Stops need not be sorted - they typically come from user configuration - and the
// caller's slice is left untouched. No stops means no colouriser: nil, which every
// primitive treats as plain white.
func NewZoneColorizer(stops []ColorStop) BarColorizer {
	if len(stops) == 0 {
		return nil
	}

	sorted := append([]ColorStop(nil), stops...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Pos < sorted[j].Pos })

	return func(pos int) color.RGBA {
		col := meterWhite
		for _, stop := range sorted {
			if pos < stop.Pos {
				break
			}
			col = stop.Col
		}
		return col
	}
}

// colorAt resolves the colour for pixel x of a bar spanning span pixels, defaulting to
// the white meters were painted with before colourisers existed.
func colorAt(colorize BarColorizer, x int, span int) color.RGBA {
	if colorize == nil || span <= 0 {
		return meterWhite
	}
	return colorize(x * 1000 / span)
}
