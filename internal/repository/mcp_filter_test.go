package repository

import (
	"reflect"
	"strings"
	"testing"
)

func TestListFilterCategoryORTransportAND(t *testing.T) {
	where, args := (ListFilter{CallerUID: "u1", SpaceID: "s1", Categories: []string{"dev", "search"}, Transports: []string{"stdio"}}).buildWhere()
	if strings.Count(where, "category IN (?,?)") != 1 || strings.Contains(where, "category = ?") {
		t.Fatalf("category predicate must be one OR set: %s", where)
	}
	if !strings.Contains(where, "transport IN (?)") {
		t.Fatalf("transport must be combined with AND: %s", where)
	}
	want := []any{"s1", "u1", "dev", "search", "stdio"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %#v, want %#v", args, want)
	}
}
