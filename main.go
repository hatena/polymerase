package main

import (
	"github.com/taku-k/polymerase/pkg/cli"
)

func main() {
	// defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	cli.Run()
}
