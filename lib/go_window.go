//go:build windows

package lib

import (
	"fmt"
	"os"

	builder "github.com/NoF0rte/cmd-builder"
)

func install(ver *VersionInfo, goArchivePath string, installPath string) error {
	fmt.Printf("[+] Installing %s\n", ver.Version)

	return builder.Cmd("msiexec", "/i", goArchivePath).
		Stdout(os.Stdout).
		Run()
}
