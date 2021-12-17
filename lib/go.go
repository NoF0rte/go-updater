package lib

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/hashicorp/go-version"
)

const (
	baseUrl            = "https://go.dev/dl/"
	defaultInstallPath = "/usr/local"
)

var (
	versionRegex        = `[0-9]+\.[0-9]+\.[0-9]+`
	goVersionFileRegex  = regexp.MustCompile(fmt.Sprintf(`go(%s)\.([a-zA-Z]+)-([a-zA-Z0-9]+).*`, versionRegex))
	GoNotInstalledError = fmt.Errorf("Go not installed")
)

type VersionInfo struct {
	Path    string
	Version *version.Version
}

func GetVersions() []*VersionInfo {
	var versions []*VersionInfo

	base, _ := url.Parse(baseUrl)
	c := colly.NewCollector()
	c.OnHTML("a.download", func(h *colly.HTMLElement) {
		name := h.Text
		if strings.Contains(name, "src") {
			return
		}

		href := h.Attr("href")
		fullPath, _ := url.Parse(fmt.Sprintf("%s://%s%s", base.Scheme, base.Host, href))

		matches := goVersionFileRegex.FindStringSubmatch(name)
		if len(matches) == 0 {
			return
		}

		versionOS := matches[2]
		architecture := matches[3]
		if versionOS != runtime.GOOS || architecture != runtime.GOARCH {
			return
		}

		v, _ := version.NewSemver(matches[1])
		versions = append(versions, &VersionInfo{
			Path:    fullPath.String(),
			Version: v,
		})
	})

	c.Visit(baseUrl)

	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Version.GreaterThanOrEqual(versions[j].Version)
	})

	return versions
}

func GetLatestVersion() *VersionInfo {
	versions := GetVersions()
	if len(versions) > 0 {
		return versions[0]
	}
	return nil
}

func GetInstalledVersion() (*VersionInfo, error) {
	goPath, err := exec.LookPath("go")
	if err != nil {
		return nil, GoNotInstalledError
	}

	goRoot := filepath.Join(goPath, "../../")

	bytes, err := os.ReadFile(filepath.Join(goRoot, "VERSION"))
	if err != nil {
		return nil, fmt.Errorf("error reading go version file: %v", err)
	}

	regex := regexp.MustCompile(fmt.Sprintf(`go(%s)`, versionRegex))
	matches := regex.FindSubmatch(bytes)
	v, err := version.NewSemver(string(matches[1]))
	if err != nil {
		return nil, err
	}

	return &VersionInfo{
		Version: v,
		Path:    goRoot,
	}, nil
}

func Install(ver *VersionInfo, path string) error {
	if path == "" {
		path = defaultInstallPath
	}

	fmt.Printf("[+] Downloading %s\n", filepath.Base(ver.Path))

	downloadPath := filepath.Join(os.TempDir(), filepath.Base(ver.Path))
	err := downloadFile(downloadPath, ver.Path)
	if err != nil {
		return err
	}

	err = os.Remove(downloadPath)
	return err
}

func downloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func removeOldInstall(path string) error {

	return nil
}
