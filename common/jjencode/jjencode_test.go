package jjencode

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func writeFile(p string, v []byte) error {
	f, err := os.OpenFile(p, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(v)
	if err != nil {
		return err
	}
	return nil
}

func readFile(p string) ([]byte, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ioutil.ReadAll(f)
}

func fileEqual(p string, v []byte) (bool, error) {
	r, err := readFile(p)
	if err != nil {
		return false, err
	}

	return bytes.Equal(v, r), nil
}

func TestJjencode(t *testing.T) {
	r, err := Jjencode("$", `alert("世界你好！" )`)
	if err != nil {
		t.Fatal(err)
	}

	if r != `$=~[];$={___:++$,$$$$:(![]+"")[$],__$:++$,$_$_:(![]+"")[$],_$_:++$,$_$$:({}+"")[$],$$_$:($[$]+"")[$],_$$:++$,$$$_:(!""+"")[$],$__:++$,$_$:++$,$$__:({}+"")[$],$$_:++$,$$$:++$,$___:++$,$__$:++$};$.$_=($.$_=$+"")[$.$_$]+($._$=$.$_[$.__$])+($.$$=($.$+"")[$.__$])+((!$)+"")[$._$$]+($.__=$.$_[$.$$_])+($.$=(!""+"")[$.__$])+($._=(!""+"")[$._$_])+$.$_[$.$_$]+$.__+$._$+$.$;$.$$=$.$+(!""+"")[$._$$]+$.__+$._+$.$+$.$$;$.$=($.___)[$.$_][$.$_];$.$($.$($.$$+"\""+$.$_$_+(![]+"")[$._$_]+$.$$$_+"\\"+$.__$+$.$$_+$._$_+$.__+"(\\\"\\"+$._+$.$__+$.$$$_+$.__$+$.$$_+"\\"+$._+$.$$$+$.$_$+$.$__+$.$$__+"\\"+$._+$.$__+$.$$$$+$.$$_+$.___+"\\"+$._+$.$_$+$.$__$+$.$$$+$.$$_$+"\\"+$._+$.$$$$+$.$$$$+$.___+$.__$+"\\\"\\"+$.$__+$.___+")"+"\"")())();` {
		t.Fatal(err)
	}
}

func TestJjencodeFile(t *testing.T) {
	tdataPath := filepath.Join("testData", "temData")
	os.MkdirAll(tdataPath, 0666)
	fname := "a.js"

	defer os.RemoveAll(tdataPath)

	err := writeFile(filepath.Join(tdataPath, fname), []byte(`alert("世界你好！" )`))
	if err != nil {
		t.Fatal(err)
	}

	err = jjencodeFile("$", filepath.Join(tdataPath, fname), filepath.Join(tdataPath, fname))
	if err != nil {
		t.Fatal(err)
	}

	if i, err := fileEqual(filepath.Join(tdataPath, fname),
		[]byte(`$=~[];$={___:++$,$$$$:(![]+"")[$],__$:++$,$_$_:(![]+"")[$],_$_:++$,$_$$:({}+"")[$],$$_$:($[$]+"")[$],_$$:++$,$$$_:(!""+"")[$],$__:++$,$_$:++$,$$__:({}+"")[$],$$_:++$,$$$:++$,$___:++$,$__$:++$};$.$_=($.$_=$+"")[$.$_$]+($._$=$.$_[$.__$])+($.$$=($.$+"")[$.__$])+((!$)+"")[$._$$]+($.__=$.$_[$.$$_])+($.$=(!""+"")[$.__$])+($._=(!""+"")[$._$_])+$.$_[$.$_$]+$.__+$._$+$.$;$.$$=$.$+(!""+"")[$._$$]+$.__+$._+$.$+$.$$;$.$=($.___)[$.$_][$.$_];$.$($.$($.$$+"\""+$.$_$_+(![]+"")[$._$_]+$.$$$_+"\\"+$.__$+$.$$_+$._$_+$.__+"(\\\"\\"+$._+$.$__+$.$$$_+$.__$+$.$$_+"\\"+$._+$.$$$+$.$_$+$.$__+$.$$__+"\\"+$._+$.$__+$.$$$$+$.$$_+$.___+"\\"+$._+$.$_$+$.$__$+$.$$$+$.$$_$+"\\"+$._+$.$$$$+$.$$$$+$.___+$.__$+"\\\"\\"+$.$__+$.___+")"+"\"")())();`)); err != nil || i != true {
		t.Fatal("b!=x")
	}
}

func TestJiencodeFiles(t *testing.T) {
	tdataPath := filepath.Join("testData", "temData")
	os.MkdirAll(tdataPath, 0666)
	defer os.RemoveAll(tdataPath)

	v := `alert("世界你好！" )`
	e := `$=~[];$={___:++$,$$$$:(![]+"")[$],__$:++$,$_$_:(![]+"")[$],_$_:++$,$_$$:({}+"")[$],$$_$:($[$]+"")[$],_$$:++$,$$$_:(!""+"")[$],$__:++$,$_$:++$,$$__:({}+"")[$],$$_:++$,$$$:++$,$___:++$,$__$:++$};$.$_=($.$_=$+"")[$.$_$]+($._$=$.$_[$.__$])+($.$$=($.$+"")[$.__$])+((!$)+"")[$._$$]+($.__=$.$_[$.$$_])+($.$=(!""+"")[$.__$])+($._=(!""+"")[$._$_])+$.$_[$.$_$]+$.__+$._$+$.$;$.$$=$.$+(!""+"")[$._$$]+$.__+$._+$.$+$.$$;$.$=($.___)[$.$_][$.$_];$.$($.$($.$$+"\""+$.$_$_+(![]+"")[$._$_]+$.$$$_+"\\"+$.__$+$.$$_+$._$_+$.__+"(\\\"\\"+$._+$.$__+$.$$$_+$.__$+$.$$_+"\\"+$._+$.$$$+$.$_$+$.$__+$.$$__+"\\"+$._+$.$__+$.$$$$+$.$$_+$.___+"\\"+$._+$.$_$+$.$__$+$.$$$+$.$$_$+"\\"+$._+$.$$$$+$.$$$$+$.___+$.__$+"\\\"\\"+$.$__+$.___+")"+"\"")())();`

	files := []struct {
		p    string
		v, e []byte
	}{
		{"a.txt", []byte(v), []byte(v)},
		{"a.js", []byte(v), []byte(e)},
		{"b.js", []byte(v), []byte(e)},
		{"c.js", []byte(v), []byte(e)},
		{"a/a.js", []byte(v), []byte(e)},
		{"a/b.js", []byte(v), []byte(e)},
	}

	for _, f := range files {
		p := filepath.Join(tdataPath, f.p)
		d := filepath.Dir(p)
		if len(d) != 0 {
			os.MkdirAll(d, 0666)
		}

		err := writeFile(p, f.v)
		if err != nil {
			t.Error(err)
		}
	}

	err := JjencodeFile("$", tdataPath, tdataPath, true)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		p := filepath.Join(tdataPath, f.p)

		i, err := fileEqual(p, f.e)
		if err != nil {
			t.Error(err)
		}

		if !i {
			t.Errorf("文件 %v 内容不正确。", p)
		}
	}

}
