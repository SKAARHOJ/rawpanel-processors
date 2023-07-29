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

	// If the state contains image file for conversion, we will do that:
	if state.Processors.GfxConv != nil {
		stateGfxConverter(state)
	}

	// If the state contains data for audio meters, we will make that happen:
	if state.Processors.AudioMeter != nil {
		stateAudioMeter(state)
	}

	// If the state contains data for strength meters, we will make that happen:
	if state.Processors.StrengthMeter != nil {
		stateStrengthMeter(state)
	}

	// If the state contains instructions for converting HWCText to HWCGfx, we will do that:
	if state.Processors.TextToGraphics != nil {
		stateTextToGraphics(state)
	}

	// Test: Will render UTF8 text
	if state.Processors.UniText != nil {
		stateUniText(state)
	}

	// Test: Will render an image with border and text in the middle that announces the resolution.
	if state.Processors.Test != nil {
		stateTest(state)
	}
}

func writeImageToHWCGFx(state *rwp.HWCState, inImg image.Image) {
	// Initialize a raw panel graphics state:
	img := rwp.HWCGfx{}
	img.W = uint32(inImg.Bounds().Dx())
	img.H = uint32(inImg.Bounds().Dy())

	// Use monoImg to create a base:
	monoImg := monogfx.MonoImg{}
	monoImg.NewImage(int(img.W), int(img.H))
	img.ImageType = rwp.HWCGfx_MONO
	img.ImageData = monoImg.GetImgSlice()

	// Set up bounds:
	imgBounds := ImageBounds{X: 0, Y: 0, W: int(img.W), H: int(img.H)}

	// Map the image onto the canvas
	RenderImageOnCanvas(&img, inImg, imgBounds, "", "", "")

	// Set the new image:
	state.HWCGfx = &img
	state.HWCText = nil
	state.Processors.GfxConv = nil
}
