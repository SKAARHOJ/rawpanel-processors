package rawpanelproc

import (
	"strings"

	rawpanellib "github.com/SKAARHOJ/rawpanel-lib"
	rwp "github.com/SKAARHOJ/rawpanel-lib/ibeam_rawpanel"
)

func stateTextToGraphics(state *rwp.HWCState) {

	am := state.Processors.TextToGraphics

	if am.W == 0 || am.W > 500 || am.H == 0 || am.H > 500 {
		return
	}

	// Initialize a raw panel graphics state:
	img := rwp.HWCGfx{}
	img.W = uint32(am.W)
	img.H = uint32(am.H)

	// Use monoImg to create a base:
	width := int(am.W)
	height := int(am.H)
	shrink := int(am.Shrink)
	border := int(am.Border)

	state.HWCText.Title = strings.ToUpper(state.HWCText.Title)
	//log.Println(hwcState.HWCText, width, height, shrink, border)
	monoImg := rawpanellib.WriteDisplayTileNew(state.HWCText, width, height, shrink, border)
	img.ImageType = rwp.HWCGfx_MONO
	img.ImageData = monoImg.GetImgSlice()

	// Set the new image:
	state.HWCGfx = &img
	state.HWCText = nil
	state.Processors.GfxConv = nil
}
