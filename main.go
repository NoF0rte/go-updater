package main

import (
	"fmt"

	"github.com/NoF0rte/go-updater/lib"
)

func main() {
	latest := lib.GetLatestVersion()
	if latest == nil {
		fmt.Println("[!] No versions for your OS and Architecture")
		return
	}

	current, err := lib.GetInstalledVersion()
	if err != nil && err != lib.GoNotInstalledError {
		fmt.Printf("[!] %v", err)
	}
	if err == lib.GoNotInstalledError {
		fmt.Println("[+] Go not installed")
		current = &lib.VersionInfo{
			Path: "",
		}
	} else if current.Version.LessThan(latest.Version) {
		fmt.Printf("[+] Upgrading %s to %s\n", current.Version, latest.Version)
	} else {
		fmt.Println("[+] Go is up to date.")
		return
	}

	err = lib.Install(latest, current.Path)
	if err != nil {
		fmt.Printf("[!] %v\n", err)
	}
}
