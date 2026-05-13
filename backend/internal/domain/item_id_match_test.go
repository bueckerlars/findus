package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"findus/backend/internal/domain"
)

func TestItemIDMatchesPolicy_Sequential(t *testing.T) {
	p := domain.ItemIDPolicy{Kind: domain.ItemIDKindSequential, Prefix: "item", Width: 4, NextSeq: 1}
	require.True(t, domain.ItemIDMatchesPolicy("item_0001", p))
	require.True(t, domain.ItemIDMatchesPolicy("item0001", p)) // legacy (no separator)
	require.False(t, domain.ItemIDMatchesPolicy("01ARZ3NDEKTSV4RRFFQ69G5FAV", p))
	require.False(t, domain.ItemIDMatchesPolicy("item001", p))
}
