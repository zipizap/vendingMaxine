package collection

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
)

// compressDir2Tgz recursively tar+gz the directoryPath and its subdirectories and inner files, and return []bytes, error.
// It includes all the subdirectories and inner files of the directoryPath (but not the directoryPath itself)
// Supports: regular files, directories, and symlinks to relative-paths
func compressDir2Tgz(directoryPath string) ([]byte, error) {
	// Create a buffer to write the tar+gzipped data to.
	buf := new(bytes.Buffer)

	// Create a new gzip writer, which will write to the buffer.
	gzWriter := gzip.NewWriter(buf)

	// Create a new tar writer, which will write to the gzip writer.
	tarWriter := tar.NewWriter(gzWriter)

	// Get the absolute path of the directory.
	absPath, err := filepath.Abs(directoryPath)
	if err != nil {
		return nil, err
	}

	// Walk through the directory and its subdirectories and add each file to the tar archive.
	err2 := filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		// Check for errors.
		if err != nil {
			return err
		}

		// If the path is the directoryPath, skip it.
		if path == absPath {
			return nil
		}

		// write tar header for this file
		var header *tar.Header
		{
			var filename_or_symlinkTarget string
			if (info.Mode() & os.ModeSymlink) != 0 {
				// its a symlink
				symlink_name_fullpath := path
				symlink_target_relative, err := os.Readlink(symlink_name_fullpath)
				if err != nil {
					return nil
				}
				filename_or_symlinkTarget = symlink_target_relative
			} else {
				// it's not a symlink: dir, regular-file, ...
				filename_or_symlinkTarget = info.Name()
			}

			header, err = tar.FileInfoHeader(info, filename_or_symlinkTarget)
			if err != nil {
				return err
			}

			// Set the name of the file to be relative to the directory.
			header.Name, err = filepath.Rel(absPath, path)
			if err != nil {
				return err
			}

			// Write the header to the tar archive.
			if err := tarWriter.WriteHeader(header); err != nil {
				return err
			}
		}

		// write tar data for this file (if necessary, symlinks and dirs dont need to)
		{
			if info.Mode().IsRegular() {
				// If the file is a regular file , open it and copy its contents to the tar archive.
				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()

				if _, err := io.Copy(tarWriter, file); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err2 != nil {
		return nil, err2
	}

	// Close the tar and gzip writers.
	if err := tarWriter.Close(); err != nil {
		return nil, err
	}
	if err := gzWriter.Close(); err != nil {
		return nil, err
	}

	// Return the bytes of the tar+gzipped data.
	return buf.Bytes(), nil
}

// extractTgz2Dir extracts a tar.gz archive from a byte slice to a destination directory
// Supports: regular files, directories, and symlinks to relative-paths
func extractTgz2Dir(srcTgzBytes []byte, destDir string) error {
	// Create a bytes.Reader from srcTgzBytes
	reader := bytes.NewReader(srcTgzBytes)

	// Create a gzip reader
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	// Create a tar reader
	tarReader := tar.NewReader(gzipReader)

	// Extract files from the archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		path := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(file, tarReader); err != nil {
				return err
			}
		case tar.TypeSymlink:
			symlink_name := path
			symlink_target_relative := header.Linkname
			err = os.Symlink(symlink_target_relative, symlink_name)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// dirOfThisBinary returns the directory path of the current binary file.
func dirOfThisBinary() (string, error) {
	// Get the absolute path of the current binary file.
	absPath, err := filepath.Abs(os.Args[0])
	if err != nil {
		return "", err
	}

	// Get the directory path of the current binary file.
	dirPath := filepath.Dir(absPath)

	return dirPath, nil
}

func jsonEscape(i string) string {
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	// Trim the beginning and trailing " character
	return string(b[1 : len(b)-1])
}

// writeToFile writes the given content to the specified file.
// It returns an error if there was any issue while writing to the file.
func writeToFile(content string, filename string) error {
	// Open the file in write mode, create it if it doesn't exist, truncate it if it does.
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the content to the file.
	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filepath string) bool {
	info, err := os.Stat(filepath)

	// os.Stat returns an error if the file does not exist
	// So if it returns an error, we return false
	if os.IsNotExist(err) {
		return false
	}

	// If it does not return an error, we check if it's not a directory
	// because os.Stat returns file info even for directories
	// If it's not a directory, we return true
	return !info.IsDir()
}
