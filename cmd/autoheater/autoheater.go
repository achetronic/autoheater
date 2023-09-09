package main

import (
	"github.com/achetronic/autoheater/internal/cmd"
	"os"
	"path/filepath"
)

// Ref: https://open-meteo.com/en/docs#latitude=28.0930127&longitude=-16.6357443&hourly=temperature_2m,relativehumidity_2m,apparent_temperature&timezone=Europe%2FLondon&forecast_days=1

func main() {
	baseName := filepath.Base(os.Args[0])

	err := cmd.NewAutoheaterCommand(baseName).Execute()
	cmd.CheckError(err)
}
