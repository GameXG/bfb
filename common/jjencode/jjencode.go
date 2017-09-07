package jjencode

import (
	"fmt"

	"io/ioutil"
	"os"

	"io"
	"path/filepath"
	"strings"

	"github.com/robertkrimen/otto"
)

func JjencodeFile(gv, srcPath, dstPath string, isRecursive bool) error {
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return err
	}

	dstPathIsDir := false
	dstInfo, dstInfoErr := os.Stat(dstPath)
	if dstInfoErr != nil {
		if strings.HasSuffix(dstPath, "\\") || strings.HasSuffix(dstPath, "/") {
			dstPathIsDir = true
		}
	} else {
		if dstInfo.IsDir() {
			dstPathIsDir = true
		}
	}

	if !srcInfo.IsDir() {
		if dstPathIsDir {
			dstPath = filepath.Join(dstPath, filepath.Base(srcPath))
		}

		return jjencodeFile(gv, srcPath, dstPath)
	}

	if dstInfoErr == nil && dstInfo.IsDir() == false {
		return fmt.Errorf("源 %v 为目录时目标 %v 不能为文件。", srcPath, dstPath)
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	for {
		objs, err := src.Readdir(100)
		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("读取目录 %v 内容失败，%v", srcPath, err)
		}
		for _, obj := range objs {
			name := obj.Name()
			srcPath := filepath.Join(srcPath, name)
			dstPath := filepath.Join(dstPath, name)
			if obj.IsDir() {
				if isRecursive {
					err = JjencodeFile(gv, dstPath, srcPath, isRecursive)
					if err != nil {
						return fmt.Errorf("加密目录 %v 到 %v 失败，%v", srcPath, dstPath, err)
					}
				}
			} else {
				if strings.HasSuffix(strings.ToLower(srcPath), ".js") == true {
					err = jjencodeFile(gv, dstPath, srcPath)
					if err != nil {
						return fmt.Errorf("加密文件 %v 到 %v 失败，%v", srcPath, dstPath, err)
					}
				}
			}
		}
	}
	return nil
}

func jjencodeFile(gv, srcPath, dstPath string) error {
	srcF, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("打开 src 失败，%v", err)
	}

	b, err := ioutil.ReadAll(srcF)
	srcF.Close()
	if err != nil {
		return fmt.Errorf("读取 src 失败，%v", err)
	}

	r, err := Jjencode(gv, string(b))
	if err != nil {
		return fmt.Errorf("js加密失败，%v", err)
	}

	dstF, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		os.MkdirAll(filepath.Dir(dstPath), 0666)
		dstF, err = os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	}
	if err != nil {
		return fmt.Errorf("创建 dst 失败，%v", err)
	}
	defer dstF.Close()

	_, err = dstF.Write([]byte(r))
	if err != nil {
		return fmt.Errorf("写入加密 js 失败，%v", err)
	}

	return nil
}

func Jjencode(gv, text string) (string, error) {
	vm := otto.New()
	_, err := vm.Run(jjencodeJs)
	if err != nil {
		return "", fmt.Errorf("内部错误：js加密代码错误,%v", err)
	}

	err = vm.Set("gv", gv)
	if err != nil {
		return "", fmt.Errorf("vm.Set gv 错误，%v", err)
	}
	err = vm.Set("text", text)
	if err != nil {
		return "", fmt.Errorf("vm.Set text 错误，%v", err)
	}

	r, err := vm.Run(`result = jjencode(gv,text)`)
	if err != nil {
		return "", fmt.Errorf("js 加密错误，%v", err)
	}

	if !r.IsString() {
		return "", fmt.Errorf("返回值 %v 不是字符串", r)
	}

	t, err := r.ToString()
	if err != nil {
		return "", fmt.Errorf("无法获得字符串，%v", err)
	}

	return t, nil
}
