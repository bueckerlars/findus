package search

// RankWeights configures bm25() column weights for items_fts column order:
// name, description, template_data, additional_data, search_labels, search_location, template_type.
// Higher values emphasize matches in that column (SQLite bm25 weight semantics).
type RankWeights struct {
	Name           float64
	Description    float64
	TemplateData   float64
	AdditionalData float64
	Labels         float64
	Location       float64
	TemplateType   float64
	SubstrWeight   float64 // multiplier for trigram table bm25 contribution
	FuzzyMaxDist   int     // Levenshtein cap for fuzzy bonus (e.g. 2)
	FuzzyMinQuery  int     // min rune length of query to apply fuzzy bonus
}

func DefaultRankWeights() RankWeights {
	return RankWeights{
		Name:           12,
		Description:    4,
		TemplateData:   3,
		AdditionalData: 2.5,
		Labels:         5,
		Location:       4,
		TemplateType:   1.5,
		SubstrWeight:   0.35,
		FuzzyMaxDist:   2,
		FuzzyMinQuery:  4,
	}
}

// BM25Args returns weights in column order for SQL bm25(items_fts, ...).
func (w RankWeights) BM25Args() []float64 {
	return []float64{
		w.Name,
		w.Description,
		w.TemplateData,
		w.AdditionalData,
		w.Labels,
		w.Location,
		w.TemplateType,
	}
}
