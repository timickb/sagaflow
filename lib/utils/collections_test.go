package utils

import (
	"slices"
	"testing"
)

func TestSliceToMap(t *testing.T) {
	type user struct {
		ID   int
		Name string
	}

	tests := []struct {
		name     string
		input    []user
		expected map[int]user
	}{
		{
			name: "build map from slice",
			input: []user{
				{ID: 1, Name: "Alice"},
				{ID: 2, Name: "Bob"},
			},
			expected: map[int]user{
				1: {ID: 1, Name: "Alice"},
				2: {ID: 2, Name: "Bob"},
			},
		},
		{
			name: "last value wins on duplicate key",
			input: []user{
				{ID: 1, Name: "Alice"},
				{ID: 1, Name: "Alice Updated"},
			},
			expected: map[int]user{
				1: {ID: 1, Name: "Alice Updated"},
			},
		},
		{
			name:     "empty slice",
			input:    []user{},
			expected: map[int]user{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SliceToMap(tt.input, func(u user) (int, user) {
				return u.ID, u
			})

			if len(result) != len(tt.expected) {
				t.Fatalf("expected len=%d, got len=%d", len(tt.expected), len(result))
			}

			for k, expectedValue := range tt.expected {
				gotValue, ok := result[k]
				if !ok {
					t.Fatalf("expected key %v to exist", k)
				}
				if gotValue != expectedValue {
					t.Fatalf("expected value %v for key %v, got %v", expectedValue, k, gotValue)
				}
			}
		})
	}
}

func TestMapSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []string
	}{
		{
			name:     "map ints to strings",
			input:    []int{1, 2, 3},
			expected: []string{"num:1", "num:2", "num:3"},
		},
		{
			name:     "empty slice",
			input:    []int{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapSlice(tt.input, func(v int) string {
				return "num:" + string(rune('0'+v))
			})

			if !slices.Equal(result, tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMapToKeysSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]int
		expected []string
	}{
		{
			name: "map keys",
			input: map[string]int{
				"a": 1,
				"b": 2,
				"c": 3,
			},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "empty map",
			input:    map[string]int{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapToKeysSlice(tt.input)

			slices.Sort(result)
			slices.Sort(tt.expected)

			if !slices.Equal(result, tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		item     int
		expected bool
	}{
		{
			name:     "item exists",
			input:    []int{1, 2, 3},
			item:     2,
			expected: true,
		},
		{
			name:     "item does not exist",
			input:    []int{1, 2, 3},
			item:     4,
			expected: false,
		},
		{
			name:     "empty slice",
			input:    []int{},
			item:     1,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Contains(tt.input, tt.item)
			if result != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFind(t *testing.T) {
	type user struct {
		ID   int
		Name string
	}

	tests := []struct {
		name      string
		input     []*user
		predicate func(*user) bool
		expected  *user
	}{
		{
			name: "find existing item",
			input: []*user{
				{ID: 1, Name: "Alice"},
				{ID: 2, Name: "Bob"},
				{ID: 3, Name: "Charlie"},
			},
			predicate: func(u *user) bool {
				return u.ID == 2
			},
			expected: &user{ID: 2, Name: "Bob"},
		},
		{
			name: "not found",
			input: []*user{
				{ID: 1, Name: "Alice"},
				{ID: 2, Name: "Bob"},
			},
			predicate: func(u *user) bool {
				return u.ID == 10
			},
			expected: nil,
		},
		{
			name:  "empty slice",
			input: []*user{},
			predicate: func(u *user) bool {
				return u.ID == 1
			},
			expected: nil,
		},
		{
			name: "returns first matching item",
			input: []*user{
				{ID: 1, Name: "Alice"},
				{ID: 2, Name: "Bob"},
				{ID: 2, Name: "Bob Duplicate"},
			},
			predicate: func(u *user) bool {
				return u.ID == 2
			},
			expected: &user{ID: 2, Name: "Bob"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Find(tt.input, tt.predicate)

			if tt.expected == nil {
				if result != nil {
					t.Fatalf("expected nil, got %+v", *result)
				}
				return
			}

			if result == nil {
				t.Fatalf("expected %+v, got nil", *tt.expected)
			}

			if *result != *tt.expected {
				t.Fatalf("expected %+v, got %+v", *tt.expected, *result)
			}
		})
	}
}
