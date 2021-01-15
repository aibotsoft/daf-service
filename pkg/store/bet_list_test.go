package store

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStore_GetResults(t *testing.T) {
	got, err := s.GetResults(context.Background())
	if assert.NoError(t, err) {
		assert.NotEmpty(t, got)
		//t.Log(got)

	}
}
