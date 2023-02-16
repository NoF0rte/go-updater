package version

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/hashicorp/go-version"
)

const (
	baseUrl = "https://go.dev/dl"
)

var (
	versionRegex       = `[0-9]+\.[0-9]+(?:\.[0-9]+)?`
	ErrGoNotInstalled  = fmt.Errorf("go not installed")
	defaultInstallPath = ""
)

type VersionInfo struct {
	Path    string
	Version *version.Version
}

type VersionResponse struct {
	Version string        `json:"version"`
	Stable  bool          `json:"stable"`
	Files   []VersionFile `json:"files"`
}

type VersionFile struct {
	Filename string `json:"filename"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Version  string `json:"version"`
	SHA256   string `json:"sha256"`
	Size     int    `json:"size"`
	Kind     string `json:"kind"`
}

func GetVersions() ([]*VersionInfo, error) {
	var versions []*VersionInfo

	resp, err := http.Get(fmt.Sprintf("%s/?mode=json&include=all", baseUrl))
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var allVersions []VersionResponse
	err = json.Unmarshal(body, &allVersions)
	if err != nil {
		return nil, err
	}

	for _, ver := range allVersions {
		if !ver.Stable {
			continue
		}

		versionStr := strings.Replace(ver.Version, "go", "", -1)
		v, err := version.NewSemver(versionStr)
		if err != nil {
			return nil, err
		}

		arch := runtime.GOARCH
		if arch == "arm" {
			arch = "armv6l"
		}

		kind := "archive"
		if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
			kind = "installer"
		}

		file := ""
		for _, f := range ver.Files {
			if f.OS != runtime.GOOS || f.Arch != arch || f.Kind != kind {
				continue
			}

			file = f.Filename
			break
		}

		if file == "" {
			continue
		}

		versions = append(versions, &VersionInfo{
			Version: v,
			Path:    fmt.Sprintf("%s/%s", baseUrl, file),
		})
	}

	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Version.GreaterThanOrEqual(versions[j].Version)
	})

	return versions, nil
}

func GetLatestVersion() (*VersionInfo, error) {
	versions, err := GetVersions()
	if err != nil {
		return nil, err
	}
	if len(versions) > 0 {
		return versions[0], nil
	}
	return nil, nil
}

func GetSpecificVersion(v string) (*VersionInfo, error) {
	v = strings.TrimPrefix(v, "v")

	specificVersion, err := version.NewSemver(v)
	if err != nil {
		return nil, err
	}

	versions, err := GetVersions()
	if err != nil {
		return nil, err
	}

	for _, ver := range versions {
		if ver.Version.Equal(specificVersion) {
			return ver, nil
		}
	}

	return nil, nil
}

func GetInstalledVersion() (*VersionInfo, error) {
	goPath, err := exec.LookPath("go")
	if err != nil {
		return nil, ErrGoNotInstalled
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

	err = install(ver, downloadPath, path)
	if err != nil {
		return err
	}

	fmt.Println("[+] Cleaning up...")
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

func init() {
	switch runtime.GOOS {
	case "linux", "darwin":
		defaultInstallPath = "/usr/local/go"
	}
}
