package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"strings"

	"archive/zip"

	"path/filepath"

	"github.com/gamexg/bfb/common/bfb"
	"github.com/gamexg/bfb/common/google/chromewebstore"
	"github.com/gamexg/bfb/common/google/oauth2"
	"github.com/gamexg/bfb/common/mjson"
	. "github.com/gamexg/bfb/common/mstring"
	"github.com/gamexg/bfb/common/mzip"
)

type Config struct {
	ManifestPath string
	ExtensionDir string
}

type Files map[string]string

func (f *Files) String() string {
	return fmt.Sprintf("%#v", f)
}

func (f *Files) Set(value string) error {
	vs := strings.Split(value, "|")
	if len(vs) != 2 {
		return fmt.Errorf("%v 格式不正确。", value)
	}

	k := vs[0]
	k = filepath.ToSlash(k)
	if strings.HasPrefix(k, "/") {
		k = k[1:]
	}

	(*f)[k] = vs[1]

	return nil
}

type Strings []string

func (s *Strings) String() string {
	return fmt.Sprintf("%#v", s)
}

func (s *Strings) Set(value string) error {
	*s = append(*s, value)
	return nil
}

type StringMap map[string]bool

func NewStringMap() StringMap {
	return StringMap(make(map[string]bool))
}
func (s *StringMap) String() string {
	return fmt.Sprintf("%#v", s)
}
func (s *StringMap) Set(value string) error {
	(*s)[value] = true
	return nil
}

func main() {

	if len(os.Args) <= 1 {
		fmt.Println("请提供命令。")
		os.Exit(-1)
	}
	cmd := os.Args[1]
	os.Args = append(os.Args[:1], os.Args[2:]...)

	switch cmd {

	case "version":
		fmt.Printf("bfb version %v(%v) \r\n", bfb.Version, bfb.IVersion)
		return

	// 编译（或不编译）
	case "build":

		manifestPath := flag.String("manifest", "manifest.json", "manifest.json 路径")
		extensionDir := flag.String("srcDir", "src", "扩展原文件目录")
		dstDir := flag.String("dstDir", "dst", "目标目录")
		browser := flag.String("browser", "", "目标浏览器")
		isCompile := flag.Bool("compile", true, "是否执行编译(混淆)，相邻的被混淆脚本会合并为一个，也就是如果 skip参数为空时 backgroud、contentscripts 会被分别合并为单个文件。")
		compilationLevel := flag.String("compilationLevel", "", "混淆等级")
		warningLevel := flag.String("warningLevel", "", "警告等级")

		var backgroudSkipCompile Strings
		var contentscriptsSkipCompile Strings

		flag.Var(&backgroudSkipCompile, "backgroudSkipCompile", "backgroud 不编译的脚本")
		flag.Var(&contentscriptsSkipCompile, "contentscriptsSkipCompile", "contentscripts 不编译的脚本")

		flag.Parse()

		bt := bfb.BrowserType(0)
		switch strings.ToLower(*browser) {
		case "":
			bt = bfb.BrowserTypeNone
		case "firefox":
			bt = bfb.BrowserTypeFirefox
		case "chrome":
			bt = bfb.BrowserTypeChrome
		default:
			fmt.Println("不支持的浏览器类型。")
			os.Exit(-1)
		}

		err := bfb.BuildExtension(*manifestPath, *extensionDir, *dstDir, bt, *isCompile, *compilationLevel, *warningLevel, backgroudSkipCompile, contentscriptsSkipCompile)
		if err != nil {
			panic(err)
		}
		return

	case "zipDir":
		j := flag.String("json", "", "dir 命名参数文件")
		srcDir := flag.String("srcDir", "", "源目录")
		dstZip := flag.String("dstZip", "", "目标文件")

		flag.Parse()

		if len(*srcDir) == 0 || len(*dstZip) == 0 {
			panic(fmt.Errorf("srcDir 、 dstZip 不能为空。"))
		}

		if len(*j) != 0 {
			var v interface{}
			err := mjson.LoadFile(&v, *j)
			if err != nil {
				panic(fmt.Errorf("json LoadFile %v 失败，%v", j, err))
			}

			*srcDir, err = T(*srcDir).Format(v)
			if err != nil {
				panic(fmt.Errorf("格式化 srcDir 失败，%v", err))
			}
			*dstZip, err = T(*dstZip).Format(v)
			if err != nil {
				panic(fmt.Errorf("格式化 dstDir 失败，%v", err))
			}

		}

		f, err := os.OpenFile(*dstZip, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0664)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		z := zip.NewWriter(f)
		defer z.Close()

		err = mzip.ZipDir(*srcDir, z, nil)
		if err != nil {
			panic(err)
		}
		return

	case "zipEdit":
		// zip 内文件替换功能

		j := flag.String("json", "", "dir 命名参数文件")
		srcZip := flag.String("srcZip", "", "")
		dstZip := flag.String("dstZip", "", "")

		var fs Files
		flag.Var(&fs, "files", "需要替换的文件，格式：zip内路径|操作系统路径 。操作系统路径为空表示删除文件")

		flag.Parse()

		if len(*srcZip) == 0 || len(*dstZip) == 0 {
			panic(fmt.Errorf("srcZip 、 dstZip 不能为空。"))
		}

		if len(*j) != 0 {
			var v interface{}
			err := mjson.LoadFile(&v, *j)
			if err != nil {
				panic(fmt.Errorf("json LoadFile %v 失败，%v", j, err))
			}

			*srcZip, err = T(*srcZip).Format(v)
			if err != nil {
				panic(fmt.Errorf("格式化 srcDir 失败，%v", err))
			}
			*dstZip, err = T(*dstZip).Format(v)
			if err != nil {
				panic(fmt.Errorf("格式化 dstDir 失败，%v", err))
			}

			nfs := make(map[string]string)
			for i, v := range fs {
				ni, err := T(i).Format(v)
				if err != nil {
					panic(fmt.Errorf("格式化 %v 失败，%v", i, err))
				}

				nv, err := T(v).Format(v)
				if err != nil {
					panic(fmt.Errorf("格式化 %v 失败，%v", v, err))
				}

				nfs[ni] = nv
			}
			fs = nfs
		}

		zr, err := zip.OpenReader(*srcZip)
		if err != nil {
			panic(err)
		}

		dstFile, err := os.OpenFile(*dstZip, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
		if err != nil {
			panic(err)
		}
		defer dstFile.Close()

		zw := zip.NewWriter(dstFile)
		defer zw.Close()

		f := func(header *zip.FileHeader, r *io.ReadCloser, err error) bool {
			v, ok := fs[header.Name]
			if !ok {
				return true
			}

			delete(fs, header.Name)

			if len(v) == 0 {
				return false
			}

			f, err := os.Open(v)
			if err != nil {
				panic(err)
			}

			*r = f
			return true
		}

		for k, v := range fs {
			f, err := os.Open(v)
			if err != nil {
				panic(err)
			}

			i, err := os.Stat(v)
			if err != nil {
				panic(err)
			}

			h, err := zip.FileInfoHeader(i)
			if err != nil {
				panic(err)
			}

			h.Name = k

			w, err := zw.CreateHeader(h)
			if err != nil {
				panic(err)
			}

			_, err = io.Copy(w, f)
			if err != nil {
				panic(err)
			}
		}

		err = mzip.ZipAppend(zw, &zr.Reader, f)
		if err != nil {
			panic(err)
		}

	case "copy":
		// zip 内文件替换功能

		var srcs Strings
		j := flag.String("json", "", "命名参数文件")
		flag.Var(&srcs, "src", "源文件或目录，允许通过多个 -src 指定多个源 。如果以 / 或 \\ 结尾，那么源目录本身不会拷贝，只拷贝目录内的子目录及文件，否则会将源目录本身拷贝到目标位置。")
		dst := flag.String("dst", "", "如果存在多个源(包括单个源以 / 或 \\ 结尾)，那么源会原名拷贝到目标目录内;如果目标以 / 或 \\ 结尾，那么源会原名拷贝到目标目录内，否则会以目标名字为新名字拷贝到目标位置。")

		flag.Parse()

		if len(srcs) == 0 || len(*dst) == 0 {
			panic(fmt.Errorf("src 、 dst 不能为空。"))
		}

		if len(*j) != 0 {
			var v interface{}
			err := mjson.LoadFile(&v, *j)
			if err != nil {
				panic(fmt.Errorf("json LoadFile %v 失败，%v", j, err))
			}

			for i, src := range srcs {
				srcs[i], err = T(src).Format(v)
				if err != nil {
					panic(fmt.Errorf("格式化 src %v 失败，%v", src, err))
				}
			}
			*dst, err = T(*dst).Format(v)
			if err != nil {
				panic(fmt.Errorf("格式化 dst %v 失败，%v", *dst, err))
			}
		}

		dstIsDir := false
		if len(srcs) > 1 {
			dstIsDir = true
		}
		for _, v := range srcs {
			if strings.HasSuffix(v, "/") || strings.HasSuffix(v, "\\") {
				dstIsDir = true
				break
			}
		}
		/*
			dstInfo, err := os.Stat(*dst)
			if err == nil && dstInfo.IsDir() {
				dstIsDir = true
			}
		*/
		if strings.HasSuffix(*dst, "/") || strings.HasSuffix(*dst, "\\") {
			dstIsDir = true
		}

		for _, src := range srcs {
			// 检查源类型，确定执行拷贝方式。
			srcInfo, err := os.Stat(src)
			if err != nil {
				panic(fmt.Errorf("获得文件 %v 信息失败，%v", src, err))
			}

			if srcInfo.IsDir() {
				dstPath := *dst

				if dstIsDir && strings.HasSuffix(src, "/") == false && strings.HasSuffix(src, "\\") == false {
					// 如果源不包含 / ，那么源是 src 本身。如果 dstIsDir == true ，那么目标是 dst 下的 src.base ，否则目标是 dst
					dstPath = filepath.Join(dstPath, filepath.Base(src))

					defer func() {
						//TODO: 拷贝文件权限、修改日期
					}()
				}

				// 如果 源 包含 / ，那么源是 src 下的子目录及文件，不包含 src 本身。这时候 dstIsDir 必定是 true ，那么目标是 dst 下的 src 每个子目录。
				err = bfb.CopyDir(dstPath, src)
				if err != nil {
					panic(fmt.Errorf("拷贝目录 %v 到 %v 失败，%v", src, dstPath, err))
				}

			} else {
				if strings.HasSuffix(src, "/") || strings.HasSuffix(src, "\\") {
					panic(fmt.Errorf("路径 %v 是文件，不是目录。", src))
				}

				dstPath := *dst
				if dstIsDir {
					dstPath = filepath.Join(dstPath, filepath.Base(src))
				}

				os.MkdirAll(filepath.Dir(dstPath), 0664)

				_, err = bfb.CopyFile(dstPath, src)
				if err != nil {
					panic(fmt.Errorf("拷贝文件 %v 到 %v 失败，%v", src, dstPath, err))
				}
			}

		}

	case "codeUrl":
		clientId := flag.String("clientId", "", "clientId ")
		if len(*clientId) == 0 {
			fmt.Println("必须提供 ClientId 。")
			os.Exit(-1)
		}

		fmt.Print(oauth2.GetCodeUrl(*clientId))
		return

	case "getRefreshToken":
		clientId := flag.String("clientId", "", "clientId ")
		client_secret := flag.String("clientSecret", "", "client_secret")
		code := flag.String("code", "", "")
		proxy := flag.String("proxy", "", "支持 http、https、socks5等协议")

		if len(*clientId) == 0 ||
			len(*client_secret) == 0 ||
			len(*code) == 0 {
			fmt.Print("clientId、clientServre、code 不能为空。")
			os.Exit(-1)
		}
		_, refresh_token, err := oauth2.GetRefreshToken(*clientId, *client_secret, *code, *proxy)
		if err != nil {
			panic(err)
		}

		fmt.Print(refresh_token)
		return

	case "chromeUp":
		appId := flag.String("appId", "", "clientId ")
		clientId := flag.String("clientId", "", "clientId ")
		client_secret := flag.String("clientSecret", "", "client_secret")
		refreshToken := flag.String("refreshToken", "", "")
		proxy := flag.String("proxy", "", "")
		fp := flag.String("filepath", "src.zip", "")

		if len(*appId) == 0 || len(*fp) == 0 ||
			len(*clientId) == 0 ||
			len(*client_secret) == 0 ||
			len(*refreshToken) == 0 {
			fmt.Print("appId、clientId、clientServre、refreshToken、filepath 不能为空。")
			os.Exit(-1)
		}

		token, err := oauth2.RefreshToken(*clientId, *client_secret, *refreshToken, *proxy)
		if err != nil {
			panic(err)
		}

		err = chromewebstore.ChromeUp(*appId, token, *fp, *proxy)
		if err != nil {
			panic(err)
		}

	case "publish":
		appId := flag.String("appId", "", "clientId ")
		clientId := flag.String("clientId", "", "clientId ")
		client_secret := flag.String("clientSecret", "", "client_secret")
		refreshToken := flag.String("refreshToken", "", "")
		proxy := flag.String("proxy", "", "")

		if len(*appId) == 0 ||
			len(*clientId) == 0 ||
			len(*client_secret) == 0 ||
			len(*refreshToken) == 0 {
			fmt.Print("appId、clientId、clientServre、refreshToken 不能为空。")
			os.Exit(-1)
		}

		token, err := oauth2.RefreshToken(*clientId, *client_secret, *refreshToken, *proxy)
		if err != nil {
			panic(err)
		}

		err = chromewebstore.Publish(*appId, token, *proxy)
		if err != nil {
			panic(err)
		}

	default:
		fmt.Printf("未知命令 %v 。\r\n", os.Args[1])
		os.Exit(-1)
	}

	//提供以下功能：
	// 转换特定浏览器格式（其实只是删除浏览器不支持的字段，防止chrome弹警告）

	// 打包为 zip 文件

	// chrome 上传
	// chrome 发布
	// chrome GetRefreshToken

	// chrome 本地签名

	// 生成未混淆的版本
	// 生成混淆的版本

	// 扩展调试版本(带 debug.js 文件)
	// 非调试版本

	// 目录形式
	// 等待上传的版本
	// 自签名打包的版本

	// 已安装的模板

	// 考虑目的

	// 方便调试

	// firefox 可以直接完成整个打包过程，当然也可以提供 debug.js 的目录版本。
	// chrome 无法全部自动完成，只能够尝试 debug.js 的目录版本及等待打包的版本

}
