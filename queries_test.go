package queries

import (
	"reflect"
	"testing"
)

func TestIsReservedName(t *testing.T) {
	testCases := []struct {
		name     string
		expected bool
	}{
		{name: "MI", expected: true},
		{name: "SS", expected: true},
		{name: "not_reserved", expected: false},
		{name: "another_not_reserved", expected: false},
		{name: "", expected: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isReservedName(tc.name)
			if result != tc.expected {
				t.Errorf("isReservedName(%s) = %v; expected %v", tc.name, result, tc.expected)
			}
		})
	}
}

func TestNewQuery(t *testing.T) {
	testCases := []struct {
		name        string
		inputQuery  string
		expectedRaw string
		expectedOrd string
		expectedMap map[string]int
	}{
		{
			name:        "Test 1",
			inputQuery:  "SELECT * FROM users WHERE id = :id AND name = :name",
			expectedRaw: "SELECT * FROM users WHERE id = :id AND name = :name",
			expectedOrd: "SELECT * FROM users WHERE id = $1 AND name = $2",
			expectedMap: map[string]int{"id": 1, "name": 2},
		},
		{
			name:        "Test 2",
			inputQuery:  "INSERT INTO users (full_name, age) VALUES (:full_name, :age)",
			expectedRaw: "INSERT INTO users (full_name, age) VALUES (:full_name, :age)",
			expectedOrd: "INSERT INTO users (full_name, age) VALUES ($1, $2)",
			expectedMap: map[string]int{"full_name": 1, "age": 2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q := NewQuery(tc.inputQuery)
			if q.Raw != tc.expectedRaw {
				t.Errorf("Raw: got %s, expected %s", q.Raw, tc.expectedRaw)
			}
			if q.OrdinalQuery != tc.expectedOrd {
				t.Errorf("OrdinalQuery: got %s, expected %s", q.OrdinalQuery, tc.expectedOrd)
			}
			if !reflect.DeepEqual(q.Mapping, tc.expectedMap) {
				t.Errorf("Mapping: got %v, expected %v", q.Mapping, tc.expectedMap)
			}
		})
	}
}
