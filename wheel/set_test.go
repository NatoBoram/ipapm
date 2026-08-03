package wheel_test

import (
	"reflect"
	"sort"
	"testing"

	"github.com/NatoBoram/ipapm/wheel"
)

func TestValues(t *testing.T) {
	set := wheel.NewSet("foo", "bar")
	values := set.Values()
	sort.Strings(values)

	expected := []string{"bar", "foo"}
	if eq := reflect.DeepEqual(values, expected); !eq {
		t.Errorf("Expected %v, got %v", expected, values)
	}
}

func TestAdd(t *testing.T) {
	set := wheel.NewSet("foo", "bar")
	set.Add("bar", "baz")

	expected := wheel.Set[string]{"foo": {}, "bar": {}, "baz": {}}
	if eq := reflect.DeepEqual(set, expected); !eq {
		t.Errorf("Expected %v, got %v", expected, set)
	}
}

func TestUnion(t *testing.T) {
	set := wheel.NewSet("foo", "bar")
	set = set.Union(wheel.Set[string]{"bar": {}, "baz": {}})

	expected := wheel.Set[string]{"foo": {}, "bar": {}, "baz": {}}
	if eq := reflect.DeepEqual(set, expected); !eq {
		t.Errorf("Expected %v, got %v", expected, set)
	}
}
