package main

import (
	"fmt"
	"runtime/debug"
)

func runVersion() error {
	v := "dev"
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		v = info.Main.Version
	}
	fmt.Println(v)
	return nil
}
