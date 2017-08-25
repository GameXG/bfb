package mjson

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestFilterPath(t *testing.T) {
	p1 := []string{""}
	np1 := filterPath(p1)

	if reflect.DeepEqual(np1, []string{}) == false {
		t.Fatal("p1")
	}

	p2 := []string{"1", "2", "3"}
	if reflect.DeepEqual(p2, filterPath(p2)) == false {
		t.Fatal("p2")
	}

	p3 := []string{"", "", "2", ""}
	if reflect.DeepEqual(p3, filterPath(p3)) == false {
		t.Fatal("p3")
	}
}

func Test_Get(t *testing.T) {
	v := make(map[string]interface{})
	v_a := make(map[string]interface{})
	v_a["b"] = "bbb"
	v["a"] = v_a
	m := interface{}(v)

	r, err := Get(&m, strings.Split("a.b", "."))
	if err != nil {
		t.Fatal(err)
	}

	sv, ok := r.(string)
	if !ok {
		t.Fatal("%T != string", r)
	}

	if sv != "bbb" {
		t.Errorf("%v != aaa", sv)
	}

	_, err = Get(&m, strings.Split("a.b.c", "."))
	if err == nil {
		t.Fatal("err==nil")
	}

	v1, err := Get(&m, strings.Split("", "."))
	if err != nil {
		t.Fatal(err)
	}

	if reflect.DeepEqual(v1, m) != true {
		t.Fatal("v1!=v")
	}
}
func Test_GetList(t *testing.T) {
	v := make(map[string]interface{})
	v_a := []interface{}{"000"}
	m := interface{}(v)

	v["a"] = v_a

	r, err := Get(&m, strings.Split("a.0", "."))
	if err != nil {
		t.Fatal(err)
	}

	sv, ok := r.(string)
	if !ok {
		t.Fatal("%T != string", r)
	}

	if sv != "000" {
		t.Errorf("%v != 000", sv)
	}

	_, err = Get(&m, strings.Split("a.b.c", "."))
	if err == nil {
		t.Fatal("err==nil")
	}

	v1, err := Get(&m, strings.Split("", "."))
	if err != nil {
		t.Fatal(err)
	}

	if reflect.DeepEqual(v1, m) != true {
		t.Fatal("v1!=v")
	}
}

func Test_GetStringValue(t *testing.T) {

	v := make(map[string]interface{})
	v_a := make(map[string]interface{})
	v_a["b"] = "bbb"
	v["a"] = v_a
	m := interface{}(v)

	r, err := GetStringValue(&m, strings.Split("a.b", "."))
	if err != nil {
		t.Fatal(err)
	}

	if r != "bbb" {
		t.Fatal("r!=bbb")
	}
}

func Test_Set(t *testing.T) {
	v := make(map[string]interface{})
	v["a"] = make(map[string]interface{})
	m := interface{}(v)

	err := Set(&m, strings.Split("a.b", "."), "ccc")
	if err != nil {
		t.Fatal(err)
	}

	d, err := json.Marshal(&m)
	if err != nil {
		t.Fatal(err)
	}

	if string(d) != `{"a":{"b":"ccc"}}` {
		t.Fatal("d!=x")
	}
}

func Test_SetList(t *testing.T) {
	v := make(map[string]interface{})
	v["a"] = []interface{}{"000"}
	m := interface{}(v)

	err := Set(&m, strings.Split("a.0", "."), "ccc")
	if err != nil {
		t.Fatal(err)
	}

	d, err := json.Marshal(&m)
	if err != nil {
		t.Fatal(err)
	}

	if string(d) != `{"a":["ccc"]}` {
		t.Fatal("d!=x")
	}
}
func Test_Del(t *testing.T) {
	v := make(map[string]interface{})
	v["list"] = []interface{}{"000"}
	v["dict"] = map[string]interface{}{"key": 000}
	m := interface{}(v)

	err := Del(&m, strings.Split("list.0", "."))
	if err != nil {
		t.Fatal(err)
	}
	err = Del(&m, strings.Split("dict.key", "."))
	if err != nil {
		t.Fatal(err)
	}

	d, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	if string(d) != `{"dict":{},"list":[]}` {
		t.Fatal("%v!=x", string(d))
	}
}
