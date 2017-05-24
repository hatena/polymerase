package main

import (
	"os"

	"github.com/taku-k/polymerase/pkg/cli"
)

func main() {
	// defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	cli.Run(os.Args)
}
