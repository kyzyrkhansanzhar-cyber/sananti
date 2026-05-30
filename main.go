package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend
var assets embed.FS

func main() {
	// Create Wails App backend controller
	app := NewApp()

	// Launch high-fidelity native desktop window
	err := wails.Run(&options.App{
		Title:             "Sananti Active Deception Shield v7.0",
		Width:             1024,
		Height:            768,
		MinWidth:          800,
		MinHeight:         600,
		DisableResize:     false,
		Fullscreen:        false,
		Frameless:         false,
		StartHidden:       false,
		HideWindowOnClose: false,
		BackgroundColour:  &options.RGBA{R: 10, G: 10, B: 20, A: 255},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
		// Translucent custom titlebar for modern macOS layout
		Mac: &mac.Options{
			TitleBar: mac.TitleBarHiddenInset(),
			About: &mac.AboutInfo{
				Title:   "Sananti Security Guard",
				Message: "Enterprise Active Deception and Anti-Fraud Protection",
			},
		},
	})

	if err != nil {
		log.Fatalf("Critical: failed to start Wails Desktop Application: %v", err)
	}
}
