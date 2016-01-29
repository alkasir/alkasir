package nexus

import (
	"runtime"

	"github.com/thomasf/lg"

	"github.com/hanjos/nexus"
	"github.com/hanjos/nexus/credentials"
)

var baseDir = "as-downloads"
var nexusClient = nexus.New("https://nexus.some.domain", credentials.None)
var repoID = "alkasir-releases"

// Quickrunner for the latest archived release
func GetMasterSnapshot(cmd string) error {
	q := BuildQuery{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
		Cmd:  cmd,
	}

	latest, err := q.getMasterSnapshot("alkasir-snapshots")
	if err != nil {
		return err
	}

	_, err = q.GetBinary(latest)
	if err != nil {
		return err
	}

	err = latest.LinkBinaries()
	if err != nil {
		return err
	}

	// err = latest.Run(cmdGlob)
	// if err != nil {
	// lg.Fatal(err)
	// }
	return nil
}

// Quickrunner for the latest archived release
func QuickReleaseRunner(cmd string) {
	q := BuildQuery{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
		Cmd:  cmd,
	}

	artifacts, err := q.GetVersions()
	latest := artifacts.Latest()
	q.GetBinary(latest)
	bin, err := q.cmdGlob()
	if err != nil {
		lg.Fatal(err)
	}

	err = latest.Run(bin)
	if err != nil {
		lg.Fatal(err)
	}
}
