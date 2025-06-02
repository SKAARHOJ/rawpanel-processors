package rawpanelproc

import (
	"bytes"
	"image"
	"math"
	"regexp"
	"strconv"
	"strings"

	su "github.com/SKAARHOJ/ibeam-lib-utils"
	monogfx "github.com/SKAARHOJ/rawpanel-lib/ibeam_lib_monogfx"
	rwp "github.com/SKAARHOJ/rawpanel-lib/ibeam_rawpanel"
	"github.com/disintegration/gift"
	log "github.com/s00500/env_logger"

	_ "golang.org/x/image/webp" // Add this line to support webp decoding
)

// Converts PNG, JPEG and GIF files into raw panel data
func stateGfxConverter(state *rwp.HWCState) {

	inImg, _, err := image.Decode(bytes.NewReader(state.Processors.GfxConv.ImageData))
	if log.Should(err) {
		return
	}

	// Initialize a raw panel graphics state:
	img := rwp.HWCGfx{}
	img.W = state.Processors.GfxConv.W
	img.H = state.Processors.GfxConv.H

	// Pick up source dimensions if none were explicitly set:
	if img.W == 0 || img.W > 500 || img.H == 0 || img.H > 500 {
		img.W = uint32(inImg.Bounds().Dx())
		img.H = uint32(inImg.Bounds().Dy())
	}

	// Use monoImg to create a base:
	monoImg := monogfx.MonoImg{}
	monoImg.NewImage(int(img.W), int(img.H))

	// Set up image type:
	switch state.Processors.GfxConv.ImageType {
	case rwp.ProcGfxConverter_RGB16bit:
		img.ImageType = rwp.HWCGfx_RGB16bit
		img.ImageData = monoImg.GetImgSliceRGB()
	case rwp.ProcGfxConverter_Gray4bit:
		img.ImageType = rwp.HWCGfx_Gray4bit
		img.ImageData = monoImg.GetImgSliceGray()
	default:
		img.ImageType = rwp.HWCGfx_MONO
		img.ImageData = monoImg.GetImgSlice()
	}

	// Set up bounds:
	imgBounds := ImageBounds{X: 0, Y: 0, W: int(img.W), H: int(img.H)}

	newImage := inImg

	// Perform scaling and filtering:
	fitting := ""
	switch state.Processors.GfxConv.Scaling {
	case rwp.ProcGfxConverter_FILL:
		fitting = "Fill"
	case rwp.ProcGfxConverter_FIT:
		fitting = "Fit"
	case rwp.ProcGfxConverter_STRETCH:
		fitting = "Stretch"
	}
	if fitting != "" || state.Processors.GfxConv.Filters != "" {
		newImage = ScalingAndFilters(inImg, fitting, imgBounds.W, imgBounds.H, state.Processors.GfxConv.Filters)
	}

	// Map the image onto the canvas
	RenderImageOnCanvas(&img, newImage, imgBounds, "", "", "")

	// Set the new image:
	state.HWCGfx = &img
	state.HWCText = nil
}

func ScalingAndFilters(srcImg image.Image, fitting string, imgWidth int, imgHeight int, imageFilters string) image.Image {

	// Set up some filters:
	g := gift.New()

	// TODO: Maybe we should check dimensions of source image to make sure we don't try to scale something like 1x100 image to fit in the height.... (could be gigantic result image, out of memory and killing system). Some check on W/H ratio may be all it takes.
	switch fitting {
	case "Fit":
		// Is source smaller in both dimensions than image bounds? Then find out which edge and scale it...
		if srcImg.Bounds().Dx() < imgWidth && srcImg.Bounds().Dy() < imgHeight {
			//fmt.Println("Fit - Smaller: ", imgWidth, ">", srcImg.Bounds().Dx(), imgHeight, ">", srcImg.Bounds().Dy())
			if srcImg.Bounds().Dx()*1000/imgWidth > srcImg.Bounds().Dy()*1000/imgHeight {
				g.Add(gift.Resize(imgWidth, 0, gift.LanczosResampling))
			} else {
				g.Add(gift.Resize(0, imgHeight, gift.LanczosResampling))
			}
		} else if srcImg.Bounds().Dx() > imgWidth || srcImg.Bounds().Dy() > imgHeight { // Is source larger in any dimension than image bounds?
			//fmt.Println("Fit - Larger: ", imgWidth, "<", srcImg.Bounds().Dx(), imgHeight, "<", srcImg.Bounds().Dy())
			if srcImg.Bounds().Dx()*1000/imgWidth > srcImg.Bounds().Dy()*1000/imgHeight {
				g.Add(gift.Resize(imgWidth, 0, gift.LanczosResampling))
			} else {
				g.Add(gift.Resize(0, imgHeight, gift.LanczosResampling))
			}
		}
	case "Fill":
		// Is source smaller in any dimension than image bounds? Then find out which edge and scale it...
		if srcImg.Bounds().Dx() < imgWidth || srcImg.Bounds().Dy() < imgHeight {
			//fmt.Println("Fill - Smaller: ", imgWidth, ">", srcImg.Bounds().Dx(), imgHeight, ">", srcImg.Bounds().Dy())
			if srcImg.Bounds().Dx()*1000/imgWidth < srcImg.Bounds().Dy()*1000/imgHeight {
				g.Add(gift.Resize(imgWidth, 0, gift.LanczosResampling))
			} else {
				g.Add(gift.Resize(0, imgHeight, gift.LanczosResampling))
			}
		} else if srcImg.Bounds().Dx() > imgWidth && srcImg.Bounds().Dy() > imgHeight { // Is source larger in both dimension than image bounds?
			//fmt.Println("Fill - Larger: ", imgWidth, "<", srcImg.Bounds().Dx(), imgHeight, "<", srcImg.Bounds().Dy())
			if srcImg.Bounds().Dx()*1000/imgWidth < srcImg.Bounds().Dy()*1000/imgHeight {
				g.Add(gift.Resize(imgWidth, 0, gift.LanczosResampling))
			} else {
				g.Add(gift.Resize(0, imgHeight, gift.LanczosResampling))
			}
		}
	case "Stretch":
		if imgWidth != srcImg.Bounds().Dx() || imgHeight != srcImg.Bounds().Dy() {
			//fmt.Println("Stretch: ", imgWidth, "!=", srcImg.Bounds().Dx(), imgHeight, "!=", srcImg.Bounds().Dy())
			g.Add(gift.Resize(imgWidth, imgHeight, gift.LanczosResampling))
		}
	default:
		r1 := regexp.MustCompile(`^([0-9]+)x([0-9]+)$`)
		matches := r1.FindStringSubmatch(fitting)
		if len(matches) == 3 {
			imgWidth = su.Intval(matches[1])
			imgHeight = su.Intval(matches[2])
			if imgWidth >= 0 && imgHeight >= 0 && imgWidth < 500 && imgHeight < 500 && (imgWidth != srcImg.Bounds().Dx() || imgHeight != srcImg.Bounds().Dy()) { // Limit to 1-499 in sizes. (for now)...
				if imgHeight == 0 && imgWidth > 0 {
					g.Add(gift.ResizeToFit(imgWidth, 500, gift.LanczosResampling)) // 500 = the assumed max that it will fit within
				} else if imgWidth == 0 && imgHeight > 0 {
					g.Add(gift.ResizeToFit(500, imgHeight, gift.LanczosResampling)) // 500 = the assumed max that it will fit within
				} else if imgWidth > 0 && imgHeight > 0 {
					g.Add(gift.Resize(imgWidth, imgHeight, gift.LanczosResampling))
				}
			}
		}
	}

	if imageFilters != "" {
		filters := strings.Split(imageFilters, ",")
		for _, filter := range filters {
			filter = strings.TrimSpace(filter)
			filterParts := strings.Split(filter+"=", "=")
			filterParts[0] = strings.TrimSpace(filterParts[0])
			filterParts[1] = strings.TrimSpace(filterParts[1])
			parameterParts := strings.Split(filterParts[1]+";;", ";")
			switch filterParts[0] {
			case "Grayscale": // Grayscale
				g.Add(gift.Grayscale())
			case "FlipHorizontal": // FlipHorizontal
				g.Add(gift.FlipHorizontal())
			case "FlipVertical": // FlipVertical
				g.Add(gift.FlipVertical())
			case "Invert": // Invert
				g.Add(gift.Invert())
			case "Sharpen": // Sharpen=[0:10]
				if parameterParts[0] != "" {
					if s, err := strconv.ParseFloat(parameterParts[0], 32); err == nil {
						if s > 0 && s < 10 {
							g.Add(gift.UnsharpMask(float32(s), 1, 0))
						}
					}
				}
			case "GaussianBlur": // GaussianBlur=[0:10]
				if parameterParts[0] != "" {
					if s, err := strconv.ParseFloat(parameterParts[0], 32); err == nil {
						if s > 0 && s < 10 {
							g.Add(gift.GaussianBlur(float32(s)))
						}
					}
				}
			case "Threshold": // Threshold=[0:100, default 50]
				amount := su.ConstrainValue(su.Intval(parameterParts[0]), 0, 100)
				if parameterParts[0] == "" {
					amount = 50
				}
				g.Add(gift.Threshold(float32(amount)))
			case "Saturation": // Saturation=[-100:500]
				if parameterParts[0] != "" {
					amount := su.ConstrainValue(su.Intval(parameterParts[0]), -100, 500)
					g.Add(gift.Saturation(float32(amount)))
				}
			case "Contrast": // Contrast=[-100:100, default 0]
				amount := su.ConstrainValue(su.Intval(parameterParts[0]), -100, 100)
				if amount != 0 {
					g.Add(gift.Contrast(float32(amount)))
				}
			case "Brightness": // Brightness=[-100:100, default 0]
				amount := su.ConstrainValue(su.Intval(parameterParts[0]), -100, 100)
				if amount != 0 {
					g.Add(gift.Brightness(float32(amount)))
				}
			case "Gamma": // Gamma=[0.0:2.0, default 1]
				if parameterParts[0] != "" {
					if s, err := strconv.ParseFloat(parameterParts[0], 32); err == nil {
						if s > 0 && s < 2 {
							g.Add(gift.Gamma(float32(s)))
						}
					}
				}
			case "Colorize": // Colorize=[Hue 0:360];[Saturation 0:100];[Percentage 0:100]
				hue := su.ConstrainValue(su.Intval(parameterParts[0]), 0, 360)
				saturation := su.ConstrainValue(su.Intval(parameterParts[1]), 0, 100)
				amount := su.ConstrainValue(su.Intval(parameterParts[2]), 0, 100)
				if amount != 0 {
					g.Add(gift.Colorize(float32(hue), float32(saturation), float32(amount)))
				}
			case "Hue": // Hue=[-180:180]
				hue := su.ConstrainValue(su.Intval(parameterParts[0]), -180, 180)
				if hue != 0 {
					g.Add(gift.Hue(float32(hue)))
				}
			}
		}
	}

	// Create new image with the dimensions coming out of the filters, then render the image through the filters onto this:
	newImage := image.NewRGBA(g.Bounds(srcImg.Bounds()))
	g.Draw(newImage, srcImg)

	return newImage
}

type ImageBounds struct {
	X int
	Y int
	W int
	H int
}

func RenderImageOnCanvas(img *rwp.HWCGfx, srcImg image.Image, imgBounds ImageBounds, imageVerticalAlign string, imageHorizontalAlign string, blendmode string) {
	imageRect := srcImg.Bounds() // Lets assume min x and y is always zero...

	srcOffsetX := 0
	srcOffsetY := 0
	destOffsetX := 0
	destOffsetY := 0

	switch imageVerticalAlign {
	case "Top":
		// Don't touch
	case "Bottom":
		if imageRect.Bounds().Dy() > imgBounds.H {
			srcOffsetY = imageRect.Bounds().Dy() - imgBounds.H
		} else if imgBounds.H > imageRect.Bounds().Dy() {
			destOffsetY = imgBounds.H - imageRect.Bounds().Dy()
		}
	default:
		if imageRect.Bounds().Dy() > imgBounds.H {
			srcOffsetY = (imageRect.Bounds().Dy() - imgBounds.H) / 2
		} else if imgBounds.H > imageRect.Bounds().Dy() {
			destOffsetY = (imgBounds.H - imageRect.Bounds().Dy()) / 2
		}
	}

	switch imageHorizontalAlign {
	case "Left":
		// Don't touch
	case "Right":
		if imageRect.Bounds().Dx() > imgBounds.W {
			srcOffsetX = imageRect.Bounds().Dx() - imgBounds.W
		} else if imgBounds.W > imageRect.Bounds().Dx() {
			destOffsetX = imgBounds.W - imageRect.Bounds().Dx()
		}
	default:
		if imageRect.Bounds().Dx() > imgBounds.W {
			srcOffsetX = (imageRect.Bounds().Dx() - imgBounds.W) / 2
		} else if imgBounds.W > imageRect.Bounds().Dx() {
			destOffsetX = (imgBounds.W - imageRect.Bounds().Dx()) / 2
		}
	}

	// Changes in destination offset can be applied directly to the imgBounds
	if destOffsetX > 0 {
		imgBounds.X += destOffsetX
		imgBounds.W -= destOffsetX
		if imgBounds.W < 1 {
			return
		}
	}
	if destOffsetY > 0 {
		imgBounds.Y += destOffsetY
		imgBounds.H -= destOffsetY
		if imgBounds.H < 1 {
			return
		}
	}

	wInBytes := int(math.Ceil(float64(img.W) / 8))

	// Assuming image is starting in 0,0 (otherwise we need to consider Min.X and Min.Y, but here we believe they are always zero. Lazy...)
	for columns := 0; columns < imageRect.Max.X-srcOffsetX && columns < imgBounds.W; columns++ {
		for rows := 0; rows < imageRect.Max.Y-srcOffsetY && rows < imgBounds.H; rows++ {
			switch img.ImageType {
			case rwp.HWCGfx_RGB16bit:
				i := 2 * (img.W*uint32(imgBounds.Y+rows) + uint32(imgBounds.X+columns))
				if int(i+1) < len(img.ImageData) {
					// Source color:
					colR, colG, colB, colA := srcImg.At(columns+srcOffsetX, rows+srcOffsetY).RGBA()

					if blendmode == "Alpha" {
						// Remove the pre-multiplied black color from pixels in case we use the alpha channel (Otherwise we get a black fringe around edges)
						// It works with not removing it for Multiply and Screen.
						// Formular: (colR*0xFFFF-[BLACK]*(255-colA))/colA, where [BLACK] = 0
						if colA > 0 {
							colR = (colR * 0xFFFF / colA)
							colG = (colG * 0xFFFF / colA)
							colB = (colB * 0xFFFF / colA)
						}
					}

					if blendmode != "" {
						// Destination color:
						colDestR := uint32(img.ImageData[i+1]&0b11111) * 2114                                    // to 16 bit range
						colDestG := uint32(((img.ImageData[i]&0b111)<<3)|((img.ImageData[i+1]>>5)&0b111)) * 1040 // to 16 bit range
						colDestB := uint32((img.ImageData[i]>>3)&0b11111) * 2114                                 // to 16 bit range

						switch blendmode {
						case "Alpha":
							colR = (colR * colA >> 16) + (colDestR * (0xFFFF - colA) >> 16)
							colG = (colG * colA >> 16) + (colDestG * (0xFFFF - colA) >> 16)
							colB = (colB * colA >> 16) + (colDestB * (0xFFFF - colA) >> 16)
						case "Multiply":
							// a*b
							colR = (colR * colDestR >> 16)
							colG = (colG * colDestG >> 16)
							colB = (colB * colDestB >> 16)
						case "Screen":
							// 1-(1-a)*(1-b)
							colR = 0xFFFF - (((0xFFFF-colR)*(0xFFFF-colDestR))>>16)&0xFFFF
							colG = 0xFFFF - (((0xFFFF-colG)*(0xFFFF-colDestG))>>16)&0xFFFF
							colB = 0xFFFF - (((0xFFFF-colB)*(0xFFFF-colDestB))>>16)&0xFFFF
						}
					}
					colR = (colR >> 11) & 0b11111
					colG = (colG >> 10) & 0b111111
					colB = (colB >> 11) & 0b11111

					pixelColor := uint16((colB << 11) | (colG << 5) | colR)

					img.ImageData[i] = byte(pixelColor >> 8)     // MSB
					img.ImageData[i+1] = byte(pixelColor & 0xFF) // LSB
				}
			case rwp.HWCGfx_Gray4bit:
				i := (img.W*uint32(imgBounds.Y+rows) + uint32(imgBounds.X+columns)) / 2
				idx := (img.W*uint32(imgBounds.Y+rows) + uint32(imgBounds.X+columns)) % 2

				if int(i) < len(img.ImageData) {
					// Source color:
					colR, colG, colB, colA := srcImg.At(columns+srcOffsetX, rows+srcOffsetY).RGBA()

					if blendmode == "Alpha" {
						// Remove the pre-multiplied black color from pixels in case we use the alpha channel (Otherwise we get a black fringe around edges)
						// It works with not removing it for Multiply and Screen.
						// Formular: (colR*0xFFFF-[BLACK]*(255-colA))/colA, where [BLACK] = 0
						if colA > 0 {
							colR = (colR * 0xFFFF / colA)
							colG = (colG * 0xFFFF / colA)
							colB = (colB * 0xFFFF / colA)
						}
						/*
							if colA != 0xFFFF && colA != 0 {
								fmt.Printf("%5d,%5d,%5d => (%5d) %5d,%5d,%5d\n", colR>>8, colG>>8, colB>>8, colA>>8, 255-((colR*0xFFFF/colA)>>8), 102-((colG*0xFFFF/colA)>>8), 34-((colB*0xFFFF/colA)>>8))
							}
						*/
					}

					pixelColor := ((19595*colR + 38470*colG + 7471*colB + 1<<15) >> 16) & 0xFFFF // Gray pixel value (16 bit)

					if blendmode != "" {
						// Pick up destination background color:
						colDest := ((uint32(img.ImageData[i]) >> ((1 - idx) * 4)) & 0b1111) * 4369 // to 16 bit range

						switch blendmode {
						case "Alpha":
							pixelColor = (pixelColor * colA >> 16) + (colDest * (0xFFFF - colA) >> 16)
						case "Multiply":
							// a*b
							pixelColor = (pixelColor * colDest >> 16)
						case "Screen":
							// 1-(1-a)*(1-b)
							pixelColor = 0xFFFF - (((0xFFFF-pixelColor)*(0xFFFF-colDest))>>16)&0xFFFF
						}
					}
					pixelColor = (pixelColor >> 12) & 0b1111

					if idx == 0 {
						img.ImageData[i] = (img.ImageData[i] & byte(0b00001111)) | byte(pixelColor<<4)
					} else {
						img.ImageData[i] = (img.ImageData[i] & byte(0b11110000)) | byte(pixelColor)
					}
				}
			case rwp.HWCGfx_MONO:
				colR, colG, colB, _ := srcImg.At(columns+srcOffsetX, rows+srcOffsetY).RGBA()
				lum := (19595*colR + 38470*colG + 7471*colB + 1<<15) >> 16 // Gray pixel value (16 bit)
				color := lum > 32768                                       // BW pixel value (16 bit range assumed)

				x := imgBounds.X + columns
				y := imgBounds.Y + rows
				index := y*wInBytes + x/8
				if index >= 0 && index < len(img.ImageData) { // Check we are not writing out of bounds
					if color {
						img.ImageData[index] |= byte(0b1 << (7 - x%8))
					} else {
						img.ImageData[index] &= (0b1 << (7 - x%8)) ^ 0xFF
					}
				}
			}
		}
	}
}
