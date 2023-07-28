package rawpanelproc

import (
	_ "image/gif"  // Allow gifs to be loaded
	_ "image/jpeg" // Allow jpegs to be loaded
	_ "image/png"  // Allow pngs to be loaded

	rwp "github.com/SKAARHOJ/rawpanel-lib/ibeam_rawpanel"
)

func StateProcessor(state *rwp.HWCState) {

	if state.Processors == nil {
		return
	}

	// If the state contains image for conversion, we will execute that:
	if state.Processors.GfxConv != nil {
		stateGfxConverter(state)
	}
	// If the state contains data for audio meters, we will make that happen:
	if state.Processors.AudioMeter != nil {
		//		stateAudioMeter(state)
	}
}
