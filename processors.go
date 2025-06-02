package rawpanelproc

import (
	"image"

	monogfx "github.com/SKAARHOJ/rawpanel-lib/ibeam_lib_monogfx"
	rwp "github.com/SKAARHOJ/rawpanel-lib/ibeam_rawpanel"
)

func StateProcessor(state *rwp.HWCState) {

	if state.Processors == nil {
		return
	}

	switch {
	case state.Processors.GfxConv != nil:
		stateGfxConverter(state)

	case state.Processors.AudioMeter != nil:
		stateAudioMeter(state)

	case state.Processors.StrengthMeter != nil:
		stateStrengthMeter(state)

	case state.Processors.TextToGraphics != nil:
		stateTextToGraphics(state)

	case state.Processors.UniText != nil:
		stateUniText(state)

	case state.Processors.Icon != nil:
		stateIcon(state)

	case state.Processors.Test != nil:
		stateTest(state)
	}

	state.Processors = nil
}

func writeImageToHWCGFx(state *rwp.HWCState, inImg image.Image) {
	writeImageToHWCGFxWithType(state, inImg, 0)
}

func writeImageToHWCGFxWithType(state *rwp.HWCState, inImg image.Image, imageType uint32) {
	// Initialize a raw panel graphics state:
	img := rwp.HWCGfx{}
	img.W = uint32(inImg.Bounds().Dx())
	img.H = uint32(inImg.Bounds().Dy())

	// Use monoImg to create a base:
	monoImg := monogfx.MonoImg{}
	monoImg.NewImage(int(img.W), int(img.H))

	// Set up image type:
	switch imageType {
	case 1: // RGB
		img.ImageType = rwp.HWCGfx_RGB16bit
		img.ImageData = monoImg.GetImgSliceRGB()
	case 2: // Gray
		img.ImageType = rwp.HWCGfx_Gray4bit
		img.ImageData = monoImg.GetImgSliceGray()
	default: // 0
		img.ImageType = rwp.HWCGfx_MONO
		img.ImageData = monoImg.GetImgSlice()
	}

	// Set up bounds:
	imgBounds := ImageBounds{X: 0, Y: 0, W: int(img.W), H: int(img.H)}

	// Map the image onto the canvas
	RenderImageOnCanvas(&img, inImg, imgBounds, "", "", "")

	// Set the new image:
	state.HWCGfx = &img
	state.HWCText = nil
}
