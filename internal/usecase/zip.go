package usecase

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"time"
)

func zipDir(src, dest string) error {
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == src {
			return nil
		}
		rel, _ := filepath.Rel(src, path)
		if d.IsDir() {
			_, err := w.Create(rel + "/")
			return err
		}
		header := &zip.FileHeader{
			Name:   rel,
			Method: zip.Store,
		}
		header.SetModTime(time.Now())
		fw, err := w.CreateHeader(header)
		if err != nil {
			return err
		}
		fh, err := os.Open(path)
		if err != nil {
			return err
		}
		_, err = io.Copy(fw, fh)
		fh.Close()
		return err
	})
}
