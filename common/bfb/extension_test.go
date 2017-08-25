package bfb

import (
	"reflect"
	"testing"
)

func TestUniq(t *testing.T) {

	r := uniq([]string{"1", "2", "2", "3"})

	if reflect.DeepEqual(r, []string{"1", "2", "3"}) == false {
		t.Fatal("!=")
	}

}
