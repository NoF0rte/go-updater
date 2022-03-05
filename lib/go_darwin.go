//go:build darwin

package lib

import (
	"fmt"
	"os"
	"os/exec"
)

func install(ver *VersionInfo, goArchivePath string, installPath string) error {
	if _, err := os.Stat(installPath); !os.IsNotExist(err) {
		fmt.Println("[+] Removing current version")
		err = sudoExec(fmt.Sprintf("rm -r %s", installPath))
		if err != nil {
			return err
		}
	}

	fmt.Printf("[+] Installing %s\n", ver.Version)
	cmd := exec.Command("bash", "-c", fmt.Sprintf("open -W %s", goArchivePath))
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
