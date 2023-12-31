package rawpanelproc

import (
	"fmt"

	"github.com/golang/freetype/truetype"
	sync "github.com/sasha-s/go-deadlock"
	"golang.org/x/image/font"
)

// Default font sizes (second number is how many pixels to offset in negative y-direction at the default font size in order to center it correctly vertically)
var TTFFonts = map[string][]float64{
	"m5x7":      {16, 9}, // https://managore.itch.io/m5x7, Creative Commons Zero v1.0 Universal
	"SuperStar": {16, 7}, //https://www.dafont.com/superstar-2.font, 100% free, by memesbruh03
	"Unifont":   {16, 6}, // GPL, GnuUnifontFull-Pm9P
}

type FaceCache struct {
	sync.RWMutex
	content map[string]*font.Face
}

var fontCache *FaceCache // TODO: ONLY SINGLETHREAD USE

func init() {
	fontCache = new(FaceCache)
	fontCache.init()
}

func (f *FaceCache) init() { // SHould live in the top of the file actually...
	f.content = make(map[string]*font.Face)
}

func (f *FaceCache) get(id string) *font.Face {
	f.RLock()
	defer f.RUnlock()
	if font, exists := f.content[id]; exists {
		return font
	}
	return nil
}
func (f *FaceCache) set(id string, face *font.Face) {
	f.Lock()
	defer f.Unlock()
	f.content[id] = face
}

// LoadFontFace is a helper function to load the specified font file with
// the specified point size. Note that the returned `font.Face` objects
// are not thread safe and cannot be used in parallel across goroutines.
// You can usually just use the Context.LoadFontFace function instead of
// this package-level function.
func LoadFontFace(path string, points float64) (font.Face, error) {
	fontkey := fmt.Sprintf("%s-%f", path, points)
	face := fontCache.get(fontkey)
	if face == nil {
		fontBytes, err := ReadEmbeddedFileWithError(path)
		if err != nil {
			return nil, err
		}
		f, err := truetype.Parse(fontBytes)
		if err != nil {
			return nil, err
		}
		faceValue := truetype.NewFace(f, &truetype.Options{
			Size: points,
			// Hinting: font.HintingFull,
		})
		face = &faceValue
		fontCache.set(fontkey, face)
	}
	return *face, nil
}
