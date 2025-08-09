package main

import (
	"embed"
	"log"
	"os"
	goruntime "runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	iconPNG, errPNG := os.ReadFile("appicon.png")
	if errPNG != nil {
		log.Println("Предупреждение: не удалось загрузить appicon.png:", errPNG)
		iconPNG = nil
	}

	appMenu := menu.NewMenu()
	if goruntime.GOOS == "darwin" {
		appMenu.Append(menu.AppMenu())
		appMenu.Append(menu.EditMenu())
	}

	err := wails.Run(&options.App{
		Title:  "Shotgun App",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 24, G: 24, B: 27, A: 1}, // Более темный фон
		OnStartup:        app.startup,
		Bind: []interface{}{
			app, // Биндим главный оркестратор приложения
		},
		Menu: appMenu,
		Linux: &linux.Options{
			Icon: iconPNG,
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
