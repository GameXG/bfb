package bfb

import (
	"bytes"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"

	"fmt"

	"github.com/gamexg/bfb/common/compiler"
	"github.com/gamexg/bfb/common/mjson"
)

func uniq(value []string) []string {
	m := make(map[string]bool)
	r := make([]string, 0, len(value))

	for _, v := range value {
		if m[v] {
			continue
		}

		m[v] = true
		r = append(r, v)
	}

	return r
}

func toMap(value []string) (r map[string]bool) {
	r = make(map[string]bool)
	for _, v := range value {
		r[v] = true
	}
	return
}

func CopyFile(dstName, srcName string) (written int64, err error) {
	/*
		srcInfo, err := os.Stat(srcName)
		if err != nil {
			return 0, err
		}
		defer func() {
			lerr := os.Chtimes(dstName, time.Now(), srcInfo.ModTime())
			if lerr != nil {
				err = fmt.Errorf("Chtimes:%v", err)
			}

			// 在这里执行权限拷贝的原因是防止源文件所有者也无权写入。
			lerr = os.Chmod(dstName, srcInfo.Mode())
			if lerr != nil {
				err = fmt.Errorf("Chmod:%v", err)
			}
		}()*/

	src, err := os.Open(srcName)
	if err != nil {
		return
	}
	defer src.Close()
	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer dst.Close()
	return io.Copy(dst, src)
}

// 源目录内的目录及文件会被拷贝至 dstDir 目录内。

func CopyDir(dstDir string, srcDir string) error {
	dstDir = filepath.Clean(dstDir)
	srcDir = filepath.Clean(srcDir)

	return copyDir(dstDir, srcDir)
}

func copyDir(dstDir string, srcDir string) error {
	if strings.HasSuffix(srcDir, "/") == false && strings.HasSuffix(srcDir, "\\") == false {
		srcDir = srcDir + string(os.PathSeparator)
	}

	srcInfo, err := os.Stat(srcDir)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() == false {
		return fmt.Errorf("srcDir %v 不是目录。", srcDir)
	}

	os.MkdirAll(dstDir, 0664)

	src, err := os.Open(srcDir)
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
			return fmt.Errorf("读取目录 %v 内容失败，%v", srcDir, err)
		}

		for _, obj := range objs {
			name := obj.Name()
			srcPath := filepath.Join(srcDir, name)
			dstPath := filepath.Join(dstDir, name)
			if obj.IsDir() {
				err = copyDir(dstPath, srcPath)
				if err != nil {
					return fmt.Errorf("拷贝目录 %v 到 %v 失败，%v", srcPath, dstPath, err)
				}
			} else {
				_, err = CopyFile(dstPath, srcPath)
				if err != nil {
					return fmt.Errorf("文件 %v 到 %v 失败，%v", srcPath, dstPath, err)
				}
			}
		}
	}
	return nil
}

// 根据 Manifest 的内容在指定目录创建
// dstDir 可以使用 text/template 语法引用 Manidest 内容
// compiler 是否启用混淆，混淆会将所有 BackgroundScripts 文件合并为一个文件，所有 Content_scriptsJs 文件合并为一个文件。
func BuildExtension(manifestPath, extensionDir, dstDir string, browser BrowserType, isCompile bool, compilationLevel string, warningLevel string, backgroudSkipCompile, contentscriptsSkipCompile []string,externs []string) error {
	m := Manifest{}
	err := m.LoadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("LoadFile %v 失败，%v ", manifestPath, err)
	}

	tmp, err := template.New("t1").Parse(dstDir)
	if err != nil {
		return fmt.Errorf("路径模板 %v 解析失败，%v", dstDir, err)
	}

	buf := &bytes.Buffer{}
	err = tmp.Execute(buf, m.V)
	if err != nil {
		return fmt.Errorf("路径模板 %v Execute 失败，%v", dstDir, err)
	}

	dstDir = buf.String()

	bjs, err := m.GetBackgroundScripts()
	if err != nil {
		return fmt.Errorf("获得 BackgroundScripts 失败，%v", err)
	}
	cjs, err := m.GetContent_scriptsJs()
	if err != nil {
		if err != nil {
			return fmt.Errorf("获得 Content_scripts 失败，%v", err)
		}
	}

	//TODO 未处理扩展附加文件
	copyFiles := make([]string, 0)
	if isCompile {
		c := func(jss []string, jsonPath, fileNamePrefix string, skip []string) (err error) {
			skipMap := toMap(skip)
			newJss := make([]string, 0, len(jss))      // 最终 js 列表，包含编译及 skip 编译的 js
			compilerJss := make([]string, 0, len(jss)) // 待编译 js 列表

			// 将所有待编译js编译为1个js文件，并清空待编译js列表
			compilerJsFun := func(i int) error {
				if len(compilerJss) > 0 {
					newJsName := fmt.Sprintf("%v%x.js", fileNamePrefix, i)
					err = compiler.Compile(compilerJss, filepath.Join(dstDir, newJsName), warningLevel, compilationLevel,externs)
					if err != nil {
						return fmt.Errorf("编译js失败，%v", err)
					}
					newJss = append(newJss, newJsName)
					compilerJss = compilerJss[0:0]
				}
				return nil
			}

			for i, v := range jss {
				if skipMap[v] {
					err = compilerJsFun(i)
					if err != nil {
						return err
					}

					newJss = append(newJss, v)
					copyFiles = append(copyFiles, v)
				} else {
					compilerJss = append(compilerJss, filepath.Join(extensionDir,v))
				}
			}
			err = compilerJsFun(len(jss))
			if err != nil {
				return err
			}

			err = mjson.Set(&m.V, strings.Split(jsonPath, "."), newJss)
			if err != nil {
				return err
			}
			return nil
		}

		err = c(bjs, "background.scripts", "b", backgroudSkipCompile)
		if err != nil {
			return err
		}
		err = c(cjs, "content_scripts.0.js", "c", contentscriptsSkipCompile)
		if err != nil {
			return err
		}
	} else {
		copyFiles = append(bjs, cjs...)
	}

	copyFiles = uniq(copyFiles)
	for _, f := range copyFiles {
		_, err := CopyFile(filepath.Join(dstDir, f), filepath.Join(extensionDir, f))
		if err != nil {
			return err
		}
	}

	f, err := os.OpenFile(filepath.Join(dstDir, "manifest.json"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	return m.Save(f, browser)
}
