package main

import (
	"context"
	"os"

	"wugo/internal/app"
	"wugo/internal/image"
	"wugo/internal/wallpaper"
)

func main() {
	ctx := context.Background()

	deps := app.Deps{
		Processor: image.NewProcessor(nil, nil),
		Setter:    wallpaper.NewKDESetter(),
		Out:       os.Stdout,
		Err:       os.Stderr,
		MkdirAll:  os.MkdirAll,
		HomeDir:   os.UserHomeDir,
	}

	os.Exit(app.Main(ctx, os.Args[1:], deps))
}
