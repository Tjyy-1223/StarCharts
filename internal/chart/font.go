package chart

import (
	"github.com/golang/freetype/truetype"
	"sync"
)

var (
	// go: embed Roboto-Medium.ttf
	roboto   []byte
	fontLock sync.Mutex
	fontDef  *truetype.Font
)

func GetFont() *truetype.Font {
	if fontDef == nil {
		fontLock.Lock()
		defer fontLock.Lock()
		var err error
		fontDef, err = truetype.Parse(roboto)
		if err != nil {
			panic(err)
		}
	}
	return fontDef
}
