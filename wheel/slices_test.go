package wheel_test

import (
	"testing"

	"github.com/NatoBoram/ipapm/wheel"
)

func TestFindFunc_success(t *testing.T) {
	slice := []string{"foo", "bar"}
	found, ok := wheel.Find(slice, func(v string) bool {
		return v == "bar"
	})
	if !ok {
		t.Fatalf("expected to find \"bar\", found nothing")
	}

	if found != "bar" {
		t.Fatalf("expected to find \"bar\", found %q", found)
	}
}

func TestFindFunc_failure(t *testing.T) {
	slice := []string{"foo", "bar"}
	found, ok := wheel.Find(slice, func(v string) bool {
		return v == "baz"
	})
	if ok {
		t.Errorf("expected to find nothing, found %s", found)
	}
}
