package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/NoF0rte/go-updater/internal/version"
)

var versionOverride string
var dryRun bool

func main() {
	flag.Parse()

	var err error
	var versionToInstall *version.VersionInfo

	installLatest := versionOverride == "latest"
	if installLatest {
		versionToInstall, err = version.GetLatestVersion()
	} else {
		versionToInstall, err = version.GetSpecificVersion(versionOverride)
	}

	checkErr(err)

	if versionToInstall == nil {
		fmt.Println("[!] No versions for your OS and Architecture")
		return
	}

	current, err := version.GetInstalledVersion()
	if err != nil && err != version.ErrGoNotInstalled {
		fmt.Printf("[!] %v", err)
		return
	}

	if err == version.ErrGoNotInstalled {
		fmt.Println("[+] Go not installed")
		current = &version.VersionInfo{
			Path: "",
		}
	} else if installLatest {
		if !current.Version.LessThan(versionToInstall.Version) {
			fmt.Println("[+] Go is up to date.")
			return
		}

		logUpgrade(current, versionToInstall)
	} else {
		if current.Version.Equal(versionToInstall.Version) {
			fmt.Println("[+] Go is up to date.")
			return
		}

		if !current.Version.LessThan(versionToInstall.Version) {
			logDowngrade(current, versionToInstall)
		} else {
			logUpgrade(current, versionToInstall)
		}
	}

	if !dryRun {
		err = version.Install(versionToInstall, current.Path)
		checkErr(err)
	}
}

func logUpgrade(currentVer *version.VersionInfo, newVer *version.VersionInfo) {
	fmt.Printf("[+] Upgrading %s to %s\n", currentVer.Version, newVer.Version)
}

func logDowngrade(currentVer *version.VersionInfo, newVer *version.VersionInfo) {
	fmt.Printf("[+] Downgrading from %s to %s\n", currentVer.Version, newVer.Version)
}

func checkErr(msg interface{}) {
	if msg != nil {
		fmt.Printf("[!] %v\n", msg)
		os.Exit(1)
	}
}

func init() {
	flag.StringVar(&versionOverride, "version", "latest", "The version to install.")
	flag.BoolVar(&dryRun, "dry-run", false, "Don't install anything, just log messages.")
}
