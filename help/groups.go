package help

import (
	"cmp"
	"slices"
	"strconv"
)

// subGroup holds one blank-line-separated block within a section.
type subGroup[T any] struct {
	key     string
	order   int
	numeric bool
	entries []T
}

func addSubGroup[T any](subs []subGroup[T], key string, entry T) []subGroup[T] {
	for i := range subs {
		if subs[i].key == key {
			subs[i].entries = append(subs[i].entries, entry)
			return subs
		}
	}
	order, err := strconv.Atoi(key)
	return append(subs, subGroup[T]{
		key: key, order: order, numeric: err == nil, entries: []T{entry},
	})
}

// splitSubGroups anchors negative subkeys below ungrouped entries. Numeric
// subkeys sort ascending; non-numeric subkeys retain their first-seen slots.
func splitSubGroups[T any](subs []subGroup[T], content func([]T) Content) ([]Content, []Content) {
	var numeric []subGroup[T]
	for _, sub := range subs {
		if sub.numeric {
			numeric = append(numeric, sub)
		}
	}
	slices.SortStableFunc(
		numeric,
		func(a, b subGroup[T]) int { return cmp.Compare(a.order, b.order) },
	)
	var top, bottom []Content
	next := 0
	for _, sub := range subs {
		if sub.numeric {
			sub = numeric[next]
			next++
		}
		if sub.order < 0 {
			bottom = append(bottom, content(sub.entries))
		} else {
			top = append(top, content(sub.entries))
		}
	}
	return top, bottom
}
