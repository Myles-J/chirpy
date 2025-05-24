package utils_test

import (
	"os"
	"testing"

	"github.com/Myles-J/chirpy/internal/utils"
)

func TestMustGetenv(t *testing.T) {
	os.Setenv("TEST_ENV", "test")

	got := utils.MustGetenv("TEST_ENV")
	want := "test"

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
