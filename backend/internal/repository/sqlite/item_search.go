package sqlite

import (
	"context"
	"strconv"
	"strings"

	"findus/backend/internal/domain"
	"findus/backend/internal/search"
)

func bm25ItemsFTSExpr(w search.RankWeights) string {
	var b strings.Builder
	b.WriteString("bm25(items_fts")
	for _, x := range w.BM25Args() {
		b.WriteString(", ")
		b.WriteString(strconv.FormatFloat(x, 'f', 2, 64))
	}
	b.WriteByte(')')
	return b.String()
}

func (r *ItemRepo) searchItemsAdvanced(ctx context.Context, q string, limit int) ([]domain.Item, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return nil, nil
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	toks := search.Tokens(q)
	if len(toks) == 0 {
		return nil, nil
	}
	w := search.DefaultRankWeights()
	prefixMatch := search.BuildPrefixMatchOR(toks)
	if prefixMatch == "" {
		return r.searchLikeExtended(ctx, q, limit)
	}

	bm := bm25ItemsFTSExpr(w)
	cap := limit * 4
	if cap < 24 {
		cap = 24
	}
	if cap > 200 {
		cap = 200
	}

	mainQ := `SELECT i.id, ` + bm + ` AS rnk
FROM items i
INNER JOIN items_fts ON i.rowid = items_fts.rowid
WHERE items_fts MATCH ?
ORDER BY rnk ASC
LIMIT ?`

	mainRows, err := r.db.QueryContext(ctx, mainQ, prefixMatch, cap)
	if err != nil {
		return r.searchLikeExtended(ctx, q, limit)
	}
	mainScores := make(map[string]float64)
	for mainRows.Next() {
		var id string
		var rnk float64
		if err := mainRows.Scan(&id, &rnk); err != nil {
			_ = mainRows.Close()
			return nil, err
		}
		mainScores[id] = rnk
	}
	if err := mainRows.Close(); err != nil {
		return nil, err
	}

	subScores := make(map[string]float64)
	subMatch := search.BuildTrigramMatchOR(toks, 3)
	if subMatch != "" {
		var has int
		if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='items_fts_substr'`).Scan(&has); err == nil && has == 1 {
			subQ := `SELECT i.id, bm25(items_fts_substr) AS rnk
FROM items i
INNER JOIN items_fts_substr ON i.rowid = items_fts_substr.rowid
WHERE items_fts_substr MATCH ?
ORDER BY rnk ASC
LIMIT ?`
			sr, err := r.db.QueryContext(ctx, subQ, subMatch, cap)
			if err == nil {
				for sr.Next() {
					var id string
					var rnk float64
					if err := sr.Scan(&id, &rnk); err != nil {
						_ = sr.Close()
						return nil, err
					}
					subScores[id] = rnk
				}
				_ = sr.Close()
			}
		}
	}

	if len(mainScores) == 0 && len(subScores) == 0 {
		return r.searchLikeExtended(ctx, q, limit)
	}

	hits := search.MergeFTSHits(mainScores, subScores, w)

	idsForNames := make([]string, 0, len(hits))
	for _, h := range hits {
		idsForNames = append(idsForNames, h.ItemID)
	}
	names, err := r.loadItemNamesByIDs(ctx, idsForNames)
	if err != nil {
		return nil, err
	}
	search.ApplyFuzzyNameBoost(hits, names, q, w)

	search.SortHitsDescending(hits)
	if len(hits) > limit {
		hits = hits[:limit]
	}
	ordered := make([]string, len(hits))
	for i := range hits {
		ordered[i] = hits[i].ItemID
	}
	return r.loadItemsByIDsInOrder(ctx, ordered)
}

func (r *ItemRepo) loadItemNamesByIDs(ctx context.Context, ids []string) (map[string]string, error) {
	out := make(map[string]string)
	if len(ids) == 0 {
		return out, nil
	}
	ph, args := placeholders(len(ids))
	rows, err := r.db.QueryContext(ctx, `SELECT id, name FROM items WHERE id IN (`+ph+`)`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		out[id] = name
	}
	return out, rows.Err()
}

func (r *ItemRepo) loadItemsByIDsInOrder(ctx context.Context, ids []string) ([]domain.Item, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	ph, args := placeholders(len(ids))
	for i := range ids {
		args[i] = ids[i]
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, description, location_id, template_type, template_data, additional_data, photo_path, qr_token, created_at, updated_at
		FROM items WHERE id IN (`+ph+`)`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	byID := make(map[string]domain.Item, len(ids))
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		byID[it.ID] = *it
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := make([]domain.Item, 0, len(ids))
	for _, id := range ids {
		if it, ok := byID[id]; ok {
			out = append(out, it)
		}
	}
	return out, nil
}

func (r *ItemRepo) searchLikeExtended(ctx context.Context, q string, limit int) ([]domain.Item, error) {
	pat := "%" + strings.ReplaceAll(q, "%", "\\%") + "%"
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT i.id, i.name, i.description, i.location_id, i.template_type, i.template_data, i.additional_data, i.photo_path, i.qr_token, i.created_at, i.updated_at
		FROM items i
		WHERE i.name LIKE ? ESCAPE '\\' OR i.description LIKE ? ESCAPE '\\'
		OR i.template_data LIKE ? ESCAPE '\\'
		OR i.additional_data LIKE ? ESCAPE '\\'
		OR i.search_labels LIKE ? ESCAPE '\\'
		OR i.search_location LIKE ? ESCAPE '\\'
		OR EXISTS (
			SELECT 1 FROM item_labels il INNER JOIN labels lb ON lb.id = il.label_id
			WHERE il.item_id = i.id AND lb.name LIKE ? ESCAPE '\\'
		)
		ORDER BY i.name LIMIT ?`,
		pat, pat, pat, pat, pat, pat, pat, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Item
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *it)
	}
	return out, rows.Err()
}

// UpdateItemSearchDenorm updates denormalized search columns; FTS triggers sync indexes.
func (r *ItemRepo) UpdateItemSearchDenorm(ctx context.Context, itemID, searchLabels, searchLocation, searchBody string) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE items SET search_labels=?, search_location=?, search_body=? WHERE id=?`,
		searchLabels, searchLocation, searchBody, itemID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return r.upsertSubstrFTS(ctx, itemID)
}

func (r *ItemRepo) upsertSubstrFTS(ctx context.Context, itemID string) error {
	var has int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='items_fts_substr'`).Scan(&has); err != nil || has == 0 {
		return nil
	}
	var rowid int64
	var body string
	err := r.db.QueryRowContext(ctx, `
		SELECT rowid,
			name || ' ' || COALESCE(description, '') || ' ' || template_data || ' ' || additional_data || ' '
				|| search_labels || ' ' || search_location || ' ' || template_type || ' ' || search_body
		FROM items WHERE id = ?`, itemID).Scan(&rowid, &body)
	if err != nil {
		return err
	}
	if _, err := r.db.ExecContext(ctx, `INSERT INTO items_fts_substr(items_fts_substr, rowid) VALUES ('delete', ?)`, rowid); err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO items_fts_substr(rowid, body) VALUES (?, ?)`, rowid, body)
	return err
}

func (r *ItemRepo) deleteSubstrFTSForItemID(ctx context.Context, itemID string) error {
	var has int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='items_fts_substr'`).Scan(&has); err != nil || has == 0 {
		return nil
	}
	var rowid int64
	err := r.db.QueryRowContext(ctx, `SELECT rowid FROM items WHERE id = ?`, itemID).Scan(&rowid)
	if err != nil {
		return nil
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO items_fts_substr(items_fts_substr, rowid) VALUES ('delete', ?)`, rowid)
	return err
}

// UpdateSearchLocationForItemsAtLocation sets search_location for all items in a location.
func (r *ItemRepo) UpdateSearchLocationForItemsAtLocation(ctx context.Context, locationID, searchLocation string) error {
	if _, err := r.db.ExecContext(ctx, `UPDATE items SET search_location=? WHERE location_id=?`, searchLocation, locationID); err != nil {
		return err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id FROM items WHERE location_id=?`, locationID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return err
		}
		if err := r.upsertSubstrFTS(ctx, id); err != nil {
			return err
		}
	}
	return rows.Err()
}

func placeholders(n int) (string, []any) {
	if n <= 0 {
		return "", nil
	}
	args := make([]any, n)
	var b strings.Builder
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteByte('?')
	}
	return b.String(), args
}
