package mzip

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// 压缩指定目录下指定的文件
// f 为过滤函数，可修改相应参数。返回 false 表示忽略不压缩这个文件。
// 如果输入了目录，也会调用 f 函数，这时 r 为[]byte{}。
// 即使发生读取错误也会调用 f 函数，如果 f 函数未补全缺失的 header 或 r ，zipFiles 函数将中断并返回错误。
func ZipFiles(dir string, names []string, dstZip io.Writer,
	f func(dir string, header *zip.FileHeader, r *io.ReadCloser, err error) bool) error {

	z := zip.NewWriter(dstZip)
	defer z.Close()

	// 去重
	m := make(map[string]bool)
	for _, n := range names {
		m[n] = true
	}

	for n, _ := range m {
		err := func() error {
			name := filepath.ToSlash(n)

			var header *zip.FileHeader
			var r io.ReadCloser
			var err error

			p := filepath.Join(dir, filepath.FromSlash(n))

			fi, err := os.Stat(p)
			if err != nil {
				header = &zip.FileHeader{
					Name:   name,
					Method: zip.Deflate,
				}
				goto f
			}

			header, err = zip.FileInfoHeader(fi)
			if err != nil {
				header = &zip.FileHeader{
					Name:   name,
					Method: zip.Deflate,
				}
				goto f
			}
			header.Name = filepath.ToSlash(header.Name)

			header.Method = zip.Deflate
			if !fi.IsDir() {
				r, err = os.Open(p)
				rc := r
				if err != nil {
					goto f
				}
				defer rc.Close()
			} else {
				r = ioutil.NopCloser(bytes.NewReader([]byte{}))
			}

		f:
			if f == nil || f(dir, header, &r, err) {
				if r == nil {
					return fmt.Errorf("未设置 %v 文件的 r 。", header.Name)
				}

				defer r.Close() //可能重复关闭，不过无所谓。

				if header == nil {
					return fmt.Errorf("hrader 为 nil")
				}

				header.Name = filepath.ToSlash(header.Name)

				w, err := z.CreateHeader(header)
				if err != nil {
					return fmt.Errorf("添加文件 %v 失败，%v", name, err)
				}

				_, err = io.Copy(w, r)
				if err != nil {
					return fmt.Errorf("压缩文件 %v 失败，%v", name, err)
				}
			}
			return nil
		}()
		if err != nil {
			return err
		}
	}

	return nil
}

// 根据压缩文件生成新的压缩文件
func ZipAppend(dst *zip.Writer, src *zip.Reader, f func(header *zip.FileHeader, r *io.ReadCloser, err error) bool) error {
	defer dst.Flush()

	for _, file := range src.File {
		err := func() error {
			header := file.FileHeader
			header.CRC32 = 0

			r, err := file.Open()
			rc := r
			if err != nil {
				goto f
			}
			defer rc.Close()

		f:
			if f == nil || f(&header, &r, err) {
				if r == nil {
					return fmt.Errorf("未设置 %v 文件的 r 。", header.Name)
				}

				defer r.Close() //可能重复关闭，不过无所谓。

				header.Name = filepath.ToSlash(header.Name)

				w, err := dst.CreateHeader(&header)
				if err != nil {
					return fmt.Errorf("添加文件 %v 失败，%v", header.Name, err)
				}

				_, err = io.Copy(w, r)
				if err != nil {
					return fmt.Errorf("压缩文件 %v 失败，%v", header.Name, err)
				}
			}
			return nil
		}()

		if err != nil {
			return err
		}
	}
	return nil
}

func ZipDir(srcDir string, dstZip *zip.Writer,
	f func(dir string, header *zip.FileHeader, r *io.ReadCloser, err error) bool) error {

	srcDir = filepath.Clean(srcDir)

	defer dstZip.Flush()

	wf := func(path string, info os.FileInfo, err error) error {
		var header *zip.FileHeader
		var r io.ReadCloser

		name := strings.TrimPrefix(path, srcDir)
		name = filepath.ToSlash(name)
		if strings.HasPrefix(name, "/") {
			name = name[1:]
		}

		if len(name) == 0 {
			return nil
		}

		header, err = zip.FileInfoHeader(info)
		if err != nil {
			header = &zip.FileHeader{
				Name:   name,
				Method: zip.Deflate,
			}
			goto f
		}

		header.Method = zip.Deflate
		if !info.IsDir() {
			r, err = os.Open(path)
			rc := r
			if err != nil {
				goto f
			}
			defer rc.Close()
		} else {
			r = ioutil.NopCloser(bytes.NewReader([]byte{}))
		}

	f:
		if f == nil || f(srcDir, header, &r, err) {
			if r == nil {
				return fmt.Errorf("未设置 %v 文件的 r 。", header.Name)
			}

			defer r.Close() //可能重复关闭，不过无所谓。

			if header == nil {
				return fmt.Errorf("hrader 为 nil")
			}

			header.Name = filepath.ToSlash(header.Name)

			w, err := dstZip.CreateHeader(header)
			if err != nil {
				return fmt.Errorf("添加文件 %v 失败，%v", name, err)
			}

			_, err = io.Copy(w, r)
			if err != nil {
				return fmt.Errorf("压缩文件 %v 失败，%v", name, err)
			}
		}
		return nil
	}

	return filepath.Walk(srcDir, wf)
}
