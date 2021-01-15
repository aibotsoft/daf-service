package handler

import (
	"testing"
)

func Test_priceKey(t *testing.T) {
	got := priceKey("1", 1, 1, "h")
	t.Log(got)
}
