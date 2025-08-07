package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:         "HKCE Cloud Drive",
		Width:         1400,
		Height:        900,
		MinWidth:      1200,
		MinHeight:     800,
		MaxWidth:      2560,
		MaxHeight:     1440,
		DisableResize: false,
		Fullscreen:    false,
		Frameless:     false,
		StartHidden:   false,
		HideWindowOnClose: false,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 102, G: 126, B: 234, A: 1},
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		OnBeforeClose:    app.beforeClose,
		OnShutdown:       app.shutdown,
		WindowStartState: options.Normal,
		Bind: []interface{}{
			app,
		},
		// 启用开发者工具（仅在开发模式下）
		Debug: options.Debug{
			OpenInspectorOnStartup: false,
		},
		// 单实例应用
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               "com.hkce.cloud.drive",
			OnSecondInstanceLaunch: app.onSecondInstanceLaunch,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
