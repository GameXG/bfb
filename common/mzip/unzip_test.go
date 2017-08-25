package mzip

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestUnZip(t *testing.T) {
	dir := "Test-aasbdfktdf"
	defer func() {
		os.RemoveAll(dir)
	}()

	os.MkdirAll(dir, 0777)

	err := UnZipFile("zip_test.mzip", dir)
	if err != nil {
		t.Fatal(err)
	}

	isDir := func(path string) bool {
		fi, err := os.Stat(path)
		if err != nil {
			return false
		}
		return fi.IsDir()
	}

	for _, v := range []string{"11", "22"} {
		if isDir(filepath.Join(dir, v)) == false {
			t.Errorf("目录 %v 不存在。", v)
		}
	}

	equal := func(path string, c []byte) bool {
		f, err := os.OpenFile(path, os.O_RDONLY, 0666)
		if err != nil {
			return false
		}
		defer f.Close()

		b, err := ioutil.ReadAll(f)
		if err != nil {
			return false
		}

		return bytes.Equal(c, b)
	}

	for k, v := range map[string][]byte{
		"11\\33.txt": []byte("444"),
		"555.txt":    []byte("666"),
		"777.txt":    []byte("888"),
	} {
		if equal(filepath.Join(dir, k), v) == false {
			t.Errorf("文件 %v 内容不正确。", k)
		}
	}

}
