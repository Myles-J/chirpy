package utils_test

import (
	"testing"

	"github.com/Myles-J/chirpy/internal/utils"
)

func TestMustGetenv(t *testing.T) {
	t.Setenv("TEST_ENV", "test")

	got := utils.MustGetenv("TEST_ENV")
	want := "test"

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
