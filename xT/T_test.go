package xT

import (
	"github.com/stretchr/testify/assert"
	"path"
	"testing"
)

func TestContainsString(t *testing.T) {
	// Test Case 1: 切片中存在目标元素
	intSlice := []int{1, 2, 3, 4, 5}
	intElement := 3
	if !Contains(intSlice, intElement) {
		t.Errorf("Test Case 1 failed: Expected true, got false")
	}

	// Test Case 2: 切片中不存在目标元素
	strSlice := []string{"apple", "banana", "orange", "mango"}
	strElement := "kiwi"
	if Contains(strSlice, strElement) {
		t.Errorf("Test Case 2 failed: Expected false, got true")
	}

	// Test Case 3: 空切片
	emptySlice := []int{}
	emptyElement := 10
	if Contains(emptySlice, emptyElement) {
		t.Errorf("Test Case 3 failed: Expected false, got true")
	}
}

func TestRemove(t *testing.T) {
	cases := []struct {
		input  []string
		remove []string
		expect []string
	}{
		{
			input:  []string{"a", "b", "a", "c"},
			remove: []string{"a", "b"},
			expect: []string{"c"},
		},
		{
			input:  []string{"b", "c"},
			remove: []string{"a"},
			expect: []string{"b", "c"},
		},
		{
			input:  []string{"b", "a", "c"},
			remove: []string{"a"},
			expect: []string{"b", "c"},
		},
		{
			input:  []string{},
			remove: []string{"a"},
			expect: []string{},
		},
	}

	for _, each := range cases {
		t.Run(path.Join(each.input...), func(t *testing.T) {
			assert.ElementsMatch(t, each.expect, Remove(each.input, each.remove...))
		})
	}
}
