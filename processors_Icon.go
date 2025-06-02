package rawpanelproc

import (
	"fmt"
	"image"
	"image/color"
	"strconv"
	"strings"

	su "github.com/SKAARHOJ/ibeam-lib-utils"
	rwp "github.com/SKAARHOJ/rawpanel-lib/ibeam_rawpanel"
	"github.com/fogleman/gg"
	log "github.com/s00500/env_logger"
)

func stateIcon(state *rwp.HWCState) {
	inImg, err := generateIcon(state)
	if err != nil {
		return
	}

	writeImageToHWCGFxWithType(state, inImg, uint32(state.Processors.Icon.IconType))
}

// Converts PNG, JPEG and GIF files into raw panel data
func generateIcon(state *rwp.HWCState) (image.Image, error) {
	am := state.Processors.Icon

	if am.W == 0 || am.W > 500 || am.H == 0 || am.H > 500 {
		return nil, fmt.Errorf("Invalid dimensions")
	}

	srcImg := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{int(am.W), int(am.H)}})

	glyphSize := am.H
	if am.GlyphSize > 0 {
		glyphSize = am.GlyphSize
	}

	face, err := LoadFontFace("embedded/gfx/fonts/MaterialIcons-Regular.ttf", float64(su.ConstrainValue(int(glyphSize), 1, 200)))
	// _ = glyphSize
	// face, err := LoadFontFace("embedded/gfx/fonts/Unifont.ttf", 16)
	if log.Should(err) {
		return nil, fmt.Errorf("Invalid font file: %v", err)
	}

	dc := gg.NewContextForRGBA(srcImg)
	dc.SetFontFace(face)

	// Title:
	dc.SetColor(color.White)

	if am.Background {
		dc.DrawRoundedRectangle(0, 0, float64(am.W), float64(am.H), float64(su.ConstrainValue(int(am.BackgroundRadius), 1, 100)))
		dc.Fill()
		dc.SetColor(color.Black)
	}

	glyphString, err := ParseGlyphString(am.GlyphCode)
	if log.Should(err) {
		return nil, fmt.Errorf("Invalid glyph code: %v", err)
	}

	//glyphString = "ABC"
	strWidth, strHeight := dc.MeasureString(glyphString)
	xOffset := su.ConstrainValue((int(am.W)-int(strWidth))/2, -1000, 1000)
	yOffset := su.ConstrainValue(int(strHeight)+(int(am.H)-int(strHeight))/2, -1000, 1000)
	//log.Println("Glyph String:", glyphString, "Width:", strWidth, "Height:", strHeight, "X Offset:", xOffset, "Y Offset:", yOffset)
	dc.DrawString(glyphString, float64(xOffset), float64(yOffset))

	return srcImg, nil
}

func ParseGlyphString(input string) (string, error) {
	var codePoint int64
	var err error

	// Determine if input is hex (contains letters a–f or prefix 0x)
	if strings.HasPrefix(input, "0x") || strings.ContainsAny(input, "abcdefABCDEF") {
		codePoint, err = strconv.ParseInt(strings.TrimPrefix(input, "0x"), 16, 32)
	} else {
		codePoint, err = strconv.ParseInt(input, 10, 32)
	}

	if err != nil {
		return "", fmt.Errorf("invalid input: %v", err)
	}

	// Check valid Unicode range
	if codePoint < 0 || codePoint > 0x10FFFF {
		return "", fmt.Errorf("code point out of Unicode range: %d", codePoint)
	}

	return string(rune(codePoint)), nil
}
