package bfb

import (
	"fmt"

	"encoding/json"

	"io"

	"strings"

	"github.com/gamexg/bfb/common/mjson"
)

type Manifest struct {
	V interface{}
}

type BrowserType int

const (
	BrowserTypeNone BrowserType = iota
	BrowserTypeIe
	BrowserTypeFirefox
	BrowserTypeOpera
	BrowserTypeChrome
	BrowserTypeEdge
)

var NotFound = fmt.Errorf("Not found")

func (m *Manifest) LoadFile(path string) error {
	return mjson.LoadFile(&m.V, path)
}

func (m *Manifest) Save(w io.Writer, browser BrowserType) error {
	var v interface{}
	d, err := json.Marshal(m.V)
	if err != nil {
		return err
	}
	err = json.Unmarshal(d, &v)
	if err != nil {
		return err
	}

	switch browser {
	case BrowserTypeNone, BrowserTypeFirefox:
	case BrowserTypeChrome:
		mjson.Del(&v, []string{"applications"})
	default:
		return fmt.Errorf("不支持的浏览器格式。")
	}

	return mjson.Save(&v, w)
}

func toStrings(s []interface{}) ([]string, error) {
	r := make([]string, 0, len(s))

	for i, v := range s {
		sv, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("index:%V value:&#V type:%T 不是 string。", i, v, v)
		}
		r = append(r, sv)
	}
	return r, nil
}

func (m *Manifest) GetBackgroundScripts() ([]string, error) {
	v, err := mjson.GetSliceValue(&m.V, strings.Split("background.scripts", "."))
	if err != nil {
		return nil, err
	}
	return toStrings(v)
}

func (m *Manifest) GetContent_scriptsJs() ([]string, error) {
	v, err := mjson.GetSliceValue(&m.V, strings.Split("content_scripts.0.js", "."))
	if err != nil {
		return nil, err
	}
	return toStrings(v)
}

func (m *Manifest) GetVersion() (string, error) {
	return mjson.GetStringValue(&m.V, []string{"version"})
}

func (m *Manifest) GetName() (string, error) {
	return mjson.GetStringValue(&m.V, []string{"name"})
}
