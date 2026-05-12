package search

import "sort"

// HitScore holds merged ranking data for one item (higher Final is better).
type HitScore struct {
	ItemID     string
	MainBM25   float64 // raw bm25 from primary FTS (lower is better); 0 if missing
	SubBM25    float64 // raw bm25 from trigram FTS; 0 if missing
	HasMain    bool
	HasSub     bool
	FuzzyBonus float64
	Final      float64
}

// MergeFTSHits combines primary and substring FTS scores into a single sortable score (higher better).
func MergeFTSHits(
	main map[string]float64, // item id -> bm25 (lower better)
	sub map[string]float64,
	w RankWeights,
) []HitScore {
	ids := make(map[string]struct{})
	for id := range main {
		ids[id] = struct{}{}
	}
	for id := range sub {
		ids[id] = struct{}{}
	}
	out := make([]HitScore, 0, len(ids))
	sw := w.SubstrWeight
	if sw <= 0 {
		sw = 0.35
	}
	for id := range ids {
		h := HitScore{ItemID: id}
		if v, ok := main[id]; ok {
			h.MainBM25 = v
			h.HasMain = true
		}
		if v, ok := sub[id]; ok {
			h.SubBM25 = v
			h.HasSub = true
		}
		// bm25: lower is better → negate for "higher is better" base score
		var base float64
		if h.HasMain {
			base -= h.MainBM25
		}
		if h.HasSub {
			base -= h.SubBM25 * sw
		}
		h.Final = base
		out = append(out, h)
	}
	return out
}

// ApplyFuzzyNameBoost adds bonus to hits when query is close to item name (Levenshtein).
func ApplyFuzzyNameBoost(hits []HitScore, names map[string]string, query string, w RankWeights) {
	q := NormalizeFold(query)
	if len([]rune(q)) < w.FuzzyMinQuery {
		return
	}
	maxD := w.FuzzyMaxDist
	if maxD < 0 {
		maxD = 2
	}
	if names == nil {
		return
	}
	for i := range hits {
		name := NormalizeFold(names[hits[i].ItemID])
		if name == "" {
			continue
		}
		d := Levenshtein(q, name)
		if d <= maxD {
			// Stronger boost for exact / near-exact
			bonus := (float64(maxD+1) - float64(d)) * 8
			hits[i].FuzzyBonus = bonus
			hits[i].Final += bonus
		}
	}
}

// SortHitsDescending sorts by Final desc, then ItemID for stability.
func SortHitsDescending(hits []HitScore) {
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].Final == hits[j].Final {
			return hits[i].ItemID < hits[j].ItemID
		}
		return hits[i].Final > hits[j].Final
	})
}
