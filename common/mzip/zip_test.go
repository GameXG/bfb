package mzip

import (
	"archive/zip"
	"io"
	"testing"

	"io/ioutil"
	"strings"

	"bytes"
	"reflect"
)

func TestZipFiles(t *testing.T) {
	b := bytes.Buffer{}

	err := ZipFiles("test_data",
		[]string{"1.txt", "2", "3/3.txt", "4.txt", "5.txt"},
		&b,
		func(dir string, header *zip.FileHeader, r *io.ReadCloser, err error) bool {
			if header == nil {
				panic("header == nil")
			}
			switch header.Name {
			case "1.txt":
				header.Name = "1/1.txt"
			case "4.txt":
				// 忽略掉文件 4.txt
				return false
			case "5.txt":
				lr := ioutil.NopCloser(strings.NewReader("555"))
				*r = lr
			}
			return true
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	z, err := zip.NewReader(bytes.NewReader(b.Bytes()), int64(b.Len()))
	if err != nil {
		t.Fatal(err)
	}

	equal := func(file *zip.File, v []byte) {
		f, err := file.Open()
		if err != nil {
			t.Errorf("file:%v err:%v", file.Name, err)
		}
		defer f.Close()

		d, err := ioutil.ReadAll(f)
		if err != nil {
			t.Errorf("file:%v err:%v", file.Name, err)
		}

		if reflect.DeepEqual(v, d) == false {
			t.Errorf("file:%v err:%#v!=%#v", file.Name, v, d)
		}
	}

	for _, f := range z.File {
		switch f.Name {
		case "1/1.txt":
			equal(f, []byte("111"))
		case "2":
			if f.Mode().IsDir() == false {
				t.Errorf("file:%v 不是目录。", f.Name)
			}
			equal(f, []byte(""))
		case "3/3.txt":
			equal(f, []byte("333"))
		case "5.txt":
			equal(f, []byte("555"))
		default:
			t.Errorf("未知文件 %v", f.Name)
		}
	}
}

func TestZipFiles_nil(t *testing.T) {
	b := bytes.Buffer{}

	err := ZipFiles("test_data",
		[]string{"1.txt", "2", "3/3.txt", "4.txt"},
		&b, nil)

	if err != nil {
		t.Fatal(err)
	}

	z, err := zip.NewReader(bytes.NewReader(b.Bytes()), int64(b.Len()))
	if err != nil {
		t.Fatal(err)
	}

	equal := func(file *zip.File, v []byte) {
		f, err := file.Open()
		if err != nil {
			t.Errorf("file:%v err:%v", file.Name, err)
		}
		defer f.Close()

		d, err := ioutil.ReadAll(f)
		if err != nil {
			t.Errorf("file:%v err:%v", file.Name, err)
		}

		if reflect.DeepEqual(v, d) == false {
			t.Errorf("file:%v err:%#v!=%#v", file.Name, v, d)
		}
	}

	for _, f := range z.File {
		switch f.Name {
		case "1.txt":
			equal(f, []byte("111"))
		case "2":
			if f.Mode().IsDir() == false {
				t.Errorf("file:%v 不是目录。", f.Name)
			}
			equal(f, []byte(""))
		case "3/3.txt":
			equal(f, []byte("333"))
		case "4.txt":
			equal(f, []byte(""))
		default:
			t.Errorf("未知文件 %v", f.Name)
		}
	}
}

func TestZipAppend(t *testing.T) {
	b := bytes.Buffer{}

	err := ZipFiles("test_data",
		[]string{"1.txt", "2", "3/3.txt", "4.txt"},
		&b, nil)

	if err != nil {
		t.Fatal(err)
	}

	srcZip, err := zip.NewReader(bytes.NewReader(b.Bytes()), int64(b.Len()))
	if err != nil {
		t.Fatal(err)
	}

	dstZipBuff := bytes.Buffer{}
	dstZip := zip.NewWriter(&dstZipBuff)

	err = ZipAppend(dstZip, srcZip, func(header *zip.FileHeader, r *io.ReadCloser, err error) bool {
		switch header.Name {
		case "4.txt":
			*r = ioutil.NopCloser(strings.NewReader("444"))
		}

		return true
	})

	if err != nil {
		t.Fatal(err)
	}
	dstZip.Close()

	z, err := zip.NewReader(bytes.NewReader(dstZipBuff.Bytes()), int64(dstZipBuff.Len()))
	if err != nil {
		t.Fatal(err)
	}

	equal := func(file *zip.File, v []byte) {
		f, err := file.Open()
		if err != nil {
			t.Errorf("file:%v err:%v", file.Name, err)
		}
		defer f.Close()

		d, err := ioutil.ReadAll(f)
		if err != nil {
			t.Errorf("file:%v err:%v", file.Name, err)
		}

		if reflect.DeepEqual(v, d) == false {
			t.Errorf("file:%v err:%#v!=%#v", file.Name, v, d)
		}
	}

	for _, f := range z.File {
		switch f.Name {
		case "1.txt":
			equal(f, []byte("111"))
		case "2":
			if f.Mode().IsDir() == false {
				t.Errorf("file:%v 不是目录。", f.Name)
			}
			equal(f, []byte(""))
		case "3/3.txt":
			equal(f, []byte("333"))
		case "4.txt":
			equal(f, []byte("444"))
		default:
			t.Errorf("未知文件 %v", f.Name)
		}
	}
}
