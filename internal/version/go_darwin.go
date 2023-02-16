//go:build darwin

package version

import (
	"fmt"
	"os"

	builder "github.com/NoF0rte/cmd-builder"
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
	return builder.Shell(fmt.Sprintf("open -W %s", goArchivePath)).
		Stdout(os.Stdout).
		Run()
}
