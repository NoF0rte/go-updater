//go:build !windows

package lib

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func sudoExec(args string) error {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("sudo %s", args))
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func install(ver *VersionInfo, goArchivePath string, installPath string) error {
	if _, err := os.Stat(installPath); !os.IsNotExist(err) {
		fmt.Println("[+] Removing current version")
		err = sudoExec(fmt.Sprintf("rm -r %s", installPath))
		if err != nil {
			return err
		}
	}

	fmt.Printf("[+] Installing %s to %s\n", ver.Version, filepath.Dir(installPath))
	return sudoExec(fmt.Sprintf(`tar -C "%s" -xvf %s`, filepath.Dir(installPath), goArchivePath))
}
