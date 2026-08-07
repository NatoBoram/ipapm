package wheel_test

import (
	"reflect"
	"testing"

	"github.com/NatoBoram/ipapm/wheel"
)

func TestMergeMaps(t *testing.T) {
	first := map[string]string{
		"foo": "foo",
		"bar": "bar",
	}

	second := map[string]string{
		"bar": "bar",
		"baz": "baz",
	}

	result := wheel.MergeMaps(first, second)
	expected := map[string]string{
		"foo": "foo",
		"bar": "bar",
		"baz": "baz",
	}

	if eq := reflect.DeepEqual(result, expected); !eq {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
