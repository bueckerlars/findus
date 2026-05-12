package service

import (
	"sort"

	"findus/backend/internal/domain"
)

// pickReassignTargetID returns another template id to receive items when deleting excludeID.
func pickReassignTargetID(templates []domain.ItemTemplate, excludeID string) (string, bool) {
	type cand struct {
		order int
		id    string
	}
	var cands []cand
	for _, t := range templates {
		if t.ID == excludeID {
			continue
		}
		cands = append(cands, cand{t.SortOrder, t.ID})
	}
	if len(cands) == 0 {
		return "", false
	}
	sort.Slice(cands, func(i, j int) bool {
		if cands[i].order != cands[j].order {
			return cands[i].order < cands[j].order
		}
		return cands[i].id < cands[j].id
	})
	return cands[0].id, true
}

// ReassignTargetForDelete picks the template that receives items when deleting excludeID.
func ReassignTargetForDelete(list []domain.ItemTemplate, excludeID string) (targetID, targetDisplay string, ok bool) {
	id, ok := pickReassignTargetID(list, excludeID)
	if !ok {
		return "", "", false
	}
	for _, t := range list {
		if t.ID == id {
			return id, t.DisplayName, true
		}
	}
	return id, id, true
}
