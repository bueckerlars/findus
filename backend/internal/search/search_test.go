package search_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"findus/backend/internal/search"
)

func TestFlattenJSON(t *testing.T) {
	raw, _ := json.Marshal(map[string]any{
		"hostname": "db-01.example.com",
		"nested":   map[string]any{"k": 42, "s": "x"},
	})
	s := search.FlattenJSON(raw)
	require.Contains(t, s, "db-01.example.com")
	require.Contains(t, s, "x")
	require.Contains(t, s, "42")
}

func TestLevenshtein(t *testing.T) {
	require.Equal(t, 0, search.Levenshtein("abc", "abc"))
	require.Equal(t, 1, search.Levenshtein("abc", "abx"))
	require.Equal(t, 3, search.Levenshtein("kitten", "sitting"))
}

func TestBuildPrefixMatchOR(t *testing.T) {
	require.Equal(t, `("foo*" OR "bar*")`, search.BuildPrefixMatchOR([]string{"foo", "bar"}))
}

func TestMergeFTSHits(t *testing.T) {
	main := map[string]float64{"a": 1.0, "b": 2.0}
	sub := map[string]float64{"b": 0.5, "c": 1.0}
	h := search.MergeFTSHits(main, sub, search.DefaultRankWeights())
	search.SortHitsDescending(h)
	// higher Final first: b has both; a has only main -1; c has only sub
	require.GreaterOrEqual(t, len(h), 3)
}
