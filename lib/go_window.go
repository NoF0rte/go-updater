//go:build windows

package lib

import (
	"fmt"
	"os"
	"os/exec"
)

func install(ver *VersionInfo, goArchivePath string, installPath string) error {
	fmt.Printf("[+] Installing %s\n", ver.Version)
	cmd := exec.Command("msiexec", "/i", goArchivePath)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
