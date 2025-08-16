package main

import (
	"context"
	"embed"
	_ "embed"
	"log"
	"shotgun_code/cmd/app"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed assets/ignore.glob
var embeddedIgnoreGlob string

const defaultCustomPromptRulesContent = "no additional rules"

func main() {
	appInstance := &App{}

	err := wails.Run(&options.App{
		Title:  "Shotgun Code",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			container, err := app.NewContainer(ctx, embeddedIgnoreGlob, defaultCustomPromptRulesContent)
			if err != nil {
				log.Fatalf("Failed to create DI container: %v", err)
			}
			appInstance.startup(ctx, container)
		},
		OnDomReady: appInstance.domReady,
		OnShutdown: appInstance.shutdown,
		Bind: []interface{}{
			appInstance,
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
