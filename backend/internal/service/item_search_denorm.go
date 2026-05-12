package service

import (
	"context"
	"sort"
	"strings"

	"findus/backend/internal/domain"
	"findus/backend/internal/search"
)

func locationPathString(path []domain.LocationPathElement) string {
	if len(path) == 0 {
		return ""
	}
	parts := make([]string, len(path))
	for i, e := range path {
		parts[i] = e.Name
	}
	return strings.Join(parts, " / ")
}

func (s *Inventory) collectDescendantLocationIDs(ctx context.Context, rootID string) ([]string, error) {
	out := []string{rootID}
	for i := 0; i < len(out); i++ {
		cur := out[i]
		kids, err := s.Locations.ListChildren(ctx, &cur)
		if err != nil {
			return nil, err
		}
		for _, k := range kids {
			out = append(out, k.ID)
		}
	}
	return out, nil
}

func (s *Inventory) RefreshItemSearchDenorm(ctx context.Context, itemID string) error {
	it, err := s.Items.GetByID(ctx, itemID)
	if err != nil {
		return err
	}
	lbs, err := s.Items.ListLabelsForItem(ctx, itemID)
	if err != nil {
		return err
	}
	sort.Slice(lbs, func(i, j int) bool { return lbs[i].Name < lbs[j].Name })
	labelParts := make([]string, 0, len(lbs))
	for _, lb := range lbs {
		if n := strings.TrimSpace(lb.Name); n != "" {
			labelParts = append(labelParts, n)
		}
	}
	searchLabels := strings.Join(labelParts, " ")
	path, err := s.Locations.PathFromRoot(ctx, it.LocationID)
	if err != nil {
		return err
	}
	searchLoc := locationPathString(path)
	body := strings.TrimSpace(search.FlattenJSON(it.TemplateData) + " " + search.FlattenJSON(it.AdditionalData))
	return s.Items.UpdateItemSearchDenorm(ctx, itemID, searchLabels, searchLoc, body)
}

func (s *Inventory) refreshItemSearchDenorm(ctx context.Context, itemID string) error {
	return s.RefreshItemSearchDenorm(ctx, itemID)
}

func (s *Inventory) RefreshSearchLocationsForSubtree(ctx context.Context, rootLocationID string) error {
	return s.refreshSearchLocationsForSubtree(ctx, rootLocationID)
}

func (s *Inventory) refreshSearchLocationsForSubtree(ctx context.Context, rootLocationID string) error {
	ids, err := s.collectDescendantLocationIDs(ctx, rootLocationID)
	if err != nil {
		return err
	}
	for _, lid := range ids {
		path, err := s.Locations.PathFromRoot(ctx, lid)
		if err != nil {
			return err
		}
		locStr := locationPathString(path)
		if err := s.Items.UpdateSearchLocationForItemsAtLocation(ctx, lid, locStr); err != nil {
			return err
		}
	}
	return nil
}
