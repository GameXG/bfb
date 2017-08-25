package mjson

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func LoadFile(v *interface{}, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return Load(v, f)
}

func Load(v *interface{}, r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(v)
}

func filterPath(path []string) []string {
	if len(path) == 1 && len(path[0]) == 0 {
		return []string{}
	}
	return path
}

func Set(v *interface{}, path []string, value interface{}) error {
	path = filterPath(path)

	if len(path) == 0 {
		*v = value
		return nil
	}

	np := path[0 : len(path)-1]
	ck := strings.Join(np, ".")

	rv, err := Get(v, np)
	if err != nil {
		return err
	}

	switch t := rv.(type) {
	case map[string]interface{}:
		t[path[len(path)-1]] = value
		return nil

	case []interface{}:
		ii, err := strconv.Atoi(path[len(path)-1])
		if err != nil {
			return fmt.Errorf("key %v 路径为数组类型，%v 无法转换为数字。%v ", ck, path[len(path)-1], err)
		}
		t[ii] = value
		return nil

	default:
		return fmt.Errorf("key %v 类型错误，%v 不是字典或数组。", ck, t)
	}
}

func Del(v *interface{}, path []string) error {
	path = filterPath(path)

	if len(path) == 0 {
		*v = nil
		return nil
	}

	np := path[0 : len(path)-1]
	ck := strings.Join(np, ".")

	rv, err := Get(v, np)
	if err != nil {
		return err
	}

	switch t := rv.(type) {
	case map[string]interface{}:
		delete(t, path[len(path)-1])
		return nil

	case []interface{}:
		ii, err := strconv.Atoi(path[len(path)-1])
		if err != nil {
			return fmt.Errorf("key 路径为数组类型，%v 无法转换为数字。%v ", ck, path[len(path)-1], err)
		}

		nv := append(t[:ii], t[ii+1:len(t)]...)
		Set(v, np, nv)

		return nil

	default:
		return fmt.Errorf("key %v 类型错误，%v 不是字典或数组。", ck, t)
	}
}

func Get(v *interface{}, path []string) (interface{}, error) {
	path = filterPath(path)

	r := *v
	for i, p := range path {
		ck := strings.Join(path[:i+1], ".")
		switch t := r.(type) {
		case map[string]interface{}:
			ok := false
			r, ok = t[p]
			if !ok {
				return nil, fmt.Errorf("未找到 %v 。", ck)
			}

		case []interface{}:
			ii, err := strconv.Atoi(p)
			if err != nil {
				return nil, fmt.Errorf("key 路径为数组类型，%v 无法转换为数字。%v ", ck, p, err)
			}
			r = t[ii]

		default:
			return nil, fmt.Errorf("key %v 类型错误，%v 不是字典或数组。", ck, t)
		}

	}
	return r, nil
}

func GetStringValue(v *interface{}, path []string) (string, error) {
	lv, err := Get(v, path)
	if err != nil {
		return "", err
	}

	sv, ok := lv.(string)
	if !ok {
		return "", fmt.Errorf("path %v 的类型 %T 不是 string。", strings.Join(path, "."), v)
	}
	return sv, nil
}
func GetBoolValue(v *interface{}, path []string) (bool, error) {
	lv, err := Get(v, path)
	if err != nil {
		return false, err
	}

	sv, ok := lv.(bool)
	if !ok {
		return false, fmt.Errorf("path %v 的类型 %T 不是 bool 。", strings.Join(path, "."), v)
	}
	return sv, nil
}
func GetIntValue(v *interface{}, path []string) (int, error) {
	lv, err := Get(v, path)
	if err != nil {
		return 0, err
	}

	sv, ok := lv.(int)
	if !ok {
		return 0, fmt.Errorf("path %v 的类型 %T 不是int。", strings.Join(path, "."), v)
	}
	return sv, nil
}
func GetFloat64Value(v *interface{}, path []string) (float64, error) {
	lv, err := Get(v, path)
	if err != nil {
		return 0, err
	}

	sv, ok := lv.(float64)
	if !ok {
		return 0, fmt.Errorf("path %v 的类型 %T 不是float64。", strings.Join(path, "."), v)
	}
	return sv, nil
}
func GetMapValue(v *interface{}, path []string) (map[string]interface{}, error) {
	lv, err := Get(v, path)
	if err != nil {
		return nil, err
	}

	sv, ok := lv.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("path %v 的类型 %T map[string]interface{}。", strings.Join(path, "."), v)
	}
	return sv, nil
}

func GetSliceValue(v *interface{}, path []string) ([]interface{}, error) {
	lv, err := Get(v, path)
	if err != nil {
		return nil, err
	}

	sv, ok := lv.([]interface{})
	if !ok {
		return nil, fmt.Errorf("path %v 的类型 %T []interface{}。", strings.Join(path, "."), v)
	}
	return sv, nil
}

// 暂时未处理不同浏览器的区别
func Save(v *interface{}, w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(v)
}
