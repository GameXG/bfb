package mzip

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func UnZip(zipFile io.ReaderAt, size int64, dir string) error {
	r, err := zip.NewReader(zipFile, size)
	if err != nil {
		return err
	}

	for _, f := range r.File {
		ofilePath := filepath.Join(dir, filepath.FromSlash(f.Name))

		if f.Mode().IsDir() {
			os.MkdirAll(ofilePath, 0777)
			continue
		}
		fdir := filepath.Dir(ofilePath)
		os.MkdirAll(fdir, 0777)

		err = func() error {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			of, err := os.OpenFile(ofilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
			if err != nil {
				return err
			}
			defer of.Close()

			_, err = io.Copy(of, rc)

			return err
		}()

		if err != nil {
			return err
		}
	}

	return nil
}

func UnZipFile(zipFile string, dir string) error {
	f, err := os.Open(zipFile)
	if err != nil {
		return fmt.Errorf("打开文件 %v 失败，", zipFile, err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}
	return UnZip(f, fi.Size(), dir)
}
