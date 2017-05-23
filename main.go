package main

import (
	"os"

	"github.com/taku-k/xtralab/pkg/cli"
)

func main() {
	// defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	cli.Run(os.Args)
}
