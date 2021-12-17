package lib

import (
	"archive/tar"
	"compress/gzip"
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
	defaultInstallPath = "/usr/local/go"
	// defaultInstallPath = "/tmp/blah/go"
)

var (
	versionRegex       = `[0-9]+\.[0-9]+\.[0-9]+`
	goVersionFileRegex = regexp.MustCompile(fmt.Sprintf(`go(%s)\.([a-zA-Z]+)-([a-zA-Z0-9]+).*`, versionRegex))
	ErrGoNotInstalled  = fmt.Errorf("go not installed")
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

	fmt.Println("[+] Removing current version")
	err = os.RemoveAll(path)
	if err != nil {
		return err
	}

	fmt.Printf("[+] Installing %s to %s\n", ver.Version, filepath.Dir(path))
	err = untar(filepath.Dir(path), downloadPath)
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

func untar(dest string, file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			return nil

		// return any other error
		case err != nil:
			return err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(dest, header.Name)

		// the following switch could also be done using fi.Mode(), not sure if there
		// a benefit of using one vs. the other.
		// fi := header.FileInfo()

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()
		}
	}
}
