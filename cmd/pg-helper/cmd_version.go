package main

import (
	"fmt"
)

var (
	Version   = "v0.0.1"
	GoVersion = "1.22"
)

type VersionCmd struct {
	BuildInfo bool `help:"Print build information" default:"false"`
}

func (v *VersionCmd) Run(ctx *Context) error {
	fmt.Println("pg-helper", Version)
	if v.BuildInfo {
		fmt.Println("Built by: go", GoVersion)
	}
	return nil
}
