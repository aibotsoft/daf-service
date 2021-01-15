package handler

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHandler_collectEventsForSport(t *testing.T) {
	err := h.collectEventsForSport(context.Background(), 18)
	assert.NoError(t, err)
}
