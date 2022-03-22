package meta

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"

	"github.com/brigadecore/brigade/sdk/v3/meta"
)

// List is a generic type that represents an ordered and pageable collection.
type List[T any] struct {
	// ListMeta contains list metadata.
	ListMeta `json:"metadata"`
	// Items is a slice of items of type T.
	Items []T `json:"items,omitempty"`
}

// Len returns the length of the List's Items field.
func (l List[T]) Len() int64 {
	return int64(len(l.Items))
}

// Sort sorts the contents of a List's Items field. Because List is a generic
// type and cannot know how to compare all types, this function takes a
// comparison function as an argument. The comparison function MUST return an
// int value < 0 when its first argument is less than the second argument, 0
// when its first and second arguments are equal, and a in value > 0 when the
// first argument is greater than the second.
func (l List[T]) Sort(compare func(lhs, rhs T) int) {
	sort.Slice(l.Items, func(i, j int) bool {
		return compare(l.Items[i], l.Items[j]) < 0
	})
}

func (l List[T]) MarshalJSON() ([]byte, error) {
	kind := fmt.Sprintf("%sList", reflect.TypeOf(new(T)).Elem().Name())
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			ListMeta      `json:"metadata"`
			Items         []T `json:"items,omitempty"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       kind,
			},
			ListMeta: l.ListMeta,
			Items:    l.Items,
		},
	)
}
