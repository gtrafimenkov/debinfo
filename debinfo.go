// SPDX-License-Identifier: MIT

package debinfo

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/blakesmith/ar"
	"github.com/ulikunitz/xz"
)

// GetControlInfoFromDeb reads a deb file and returns content of
// control.tar.xz/control file
func GetControlInfoFromDeb(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := ar.NewReader(f)
	for true {
		arHeader, err := reader.Next()

		if err != nil {
			if err == io.EOF {
				return nil, errors.New("control.tar.xz is not found")
			}
			return nil, fmt.Errorf("failed to read ar archive: %v", err)
		}

		if arHeader.Name != "control.tar.xz" {
			continue
		}

		xzReader, err := xz.NewReader(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to parse control.tar.xz: %v", err)
		}

		tarReader := tar.NewReader(xzReader)
		for true {
			tarHeader, err := tarReader.Next()

			if err != nil {
				if err == io.EOF {
					return nil, errors.New("./control is not found in control.tar.xz")
				}
				return nil, fmt.Errorf("failed to read control.tar archive: %v", err)
			}

			if tarHeader.Name == "./control" {
				content, err := ioutil.ReadAll(tarReader)
				if err != nil {
					return nil, fmt.Errorf("failed to read ./control file from control.tar.xz: %v", err)
				}
				return content, nil
			}
		}
	}
	return nil, errors.New("logic error - unreachable code")
}

// ControlInfo contains parsed information from a deb control file
type ControlInfo struct {
	Package       string
	Source        string
	Version       string
	Architecture  string
	Maintainer    string
	InstalledSize int
	Provides      string
	Section       string
	Priority      string
	Homepage      string

	// We are not parsing the description at the moment
	// Description string
}

// ParseControlInfo parses information from a deb control file
func ParseControlInfo(controlInfo string) ControlInfo {
	info := ControlInfo{}
	for _, line := range strings.Split(controlInfo, "\n") {
		if len(line) == 0 {
			continue
		}
		items := strings.SplitN(line, ": ", 2)
		if len(items) != 2 {
			continue
		}

		key, value := items[0], items[1]

		switch key {
		case "Installed-Size":
			size, err := strconv.Atoi(value)
			if err == nil {
				info.InstalledSize = size
			}
		case "Package":
			info.Package = value
		case "Source":
			info.Source = value
		case "Version":
			info.Version = value
		case "Architecture":
			info.Architecture = value
		case "Maintainer":
			info.Maintainer = value
		case "Provides":
			info.Provides = value
		case "Section":
			info.Section = value
		case "Priority":
			info.Priority = value
		case "Homepage":
			info.Homepage = value
		}
	}
	return info
}
