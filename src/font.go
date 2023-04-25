package src

import (
	"os"
	"path/filepath"
)

func SetFont() {
	font := os.Getenv("FYNE_FONT")
	if len(font) != 0 {
		return
	}
	fontPath := filepath.Join(RootPath, "font", "SimSun.ttc")
	os.Setenv("FYNE_FONT", fontPath)
}
