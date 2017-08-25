package mstring

import (
	"bytes"
	"text/template"
)

type StringTemplate struct {
	tmp *template.Template
	Err error
}

func T(t string) *StringTemplate {
	f := StringTemplate{}
	f.tmp, f.Err = template.New("t1").Parse(t)
	return &f
}

func (t *StringTemplate) Format(data interface{}) (string, error) {
	if t.Err != nil {
		return "", t.Err
	}

	buf := &bytes.Buffer{}
	err := t.tmp.Execute(buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (t *StringTemplate) MustFormat(data interface{}) string {
	v, err := t.Format(data)
	if err != nil {
		panic(err)
	}
	return v
}
