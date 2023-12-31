// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fs

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CreateBuildContextArchive creates a tar archive of the given build context.
func CreateBuildContextArchive(src, dst string) error {
	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		return fmt.Errorf("dst path already exists: %s", dst)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0774); err != nil {
		return err
	}
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	switch {
	case info.IsDir():
		return TarDir(src, dst)
	case info.Mode().IsRegular():
		return TarFile(src, dst)
	default:
		return fmt.Errorf("input path %s is neither a directory nor a regular file", src)
	}
}

// ArchiveHasObject determines if the given tar archive contains the given object.
func ArchiveHasObject(archive string, path string) (bool, error) {
	reader, err := os.Open(archive)
	if err != nil {
		return false, err
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)
	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return false, err
		}
		if hdr.Name == path {
			return true, nil
		}
	}
	return false, nil
}

// ReadObjectFromArchive reads the given object in the given tar archive.
func ReadObjectFromArchive(archive string, path string) ([]byte, error) {
	reader, err := os.Open(archive)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)
	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		if hdr.Name == path {
			content := make([]byte, hdr.Size)
			tarReader.Read(content)
			return content, nil
		}
	}
	return nil, fmt.Errorf("could not find file %q in build context", path)
}
