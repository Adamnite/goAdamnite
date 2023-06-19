package threadSafeSlice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSingleThread(t *testing.T) {
	tss := NewThreadSafeSlice()
	items := []byte("Hello World!")
	for i, char := range items {
		tss.Append(char)
		assert.Equal(t, char, tss.Get(i), "get did not return same as most recently appended character")
		assert.Equal(t, tss.Get(-1), tss.Get(i), "-1 did not equal last index")
	}
	tss.ForEach(
		func(i int, val interface{}) bool {
			assert.Equal(
				t, items[i], val,
				"for each did not return correct values at index",
			)
			return true
		},
	)
	assert.Equal(
		t, len(items), tss.Len(),
		".Len() returned wrong number",
	)
	tss.Set(-1, '?')
	assert.Equal(
		t, len(items), tss.Len(),
		".Len() returned wrong number",
	)
	ans := tss.Pop(-1)
	assert.Equal(
		t, len(items)-1, tss.Len(),
		".Len() returned wrong number",
	)
	assert.Equal(
		t, '?', ans,
		"wrong answer returned popping last character",
	)
	assert.Panics(
		t, func() { tss.RemoveFrom(5, 0) },
		"did not panic with start being larger than end",
	)
	tss.RemoveFrom(0, -1) //clear out the list
	assert.Equal(
		t, 0, tss.Len(),
		"removing all items from array did not work",
	)
}
