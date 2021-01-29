package collections

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SimpleLRU(t *testing.T) {
	callback := func(key string, val interface{}) {
		fmt.Printf("evivt key %s val %v\n", key, val)
	}

	simple := NewSimpleLRUWithEvict(5, callback)
	testLRU(t, simple)

	lru := NewLRUWithEvict(5, callback)
	testLRU(t, lru)
}

func testLRU(t *testing.T, lru LRUInterface) {
	as := assert.New(t)

	for i := 0; i < 10; i++ {
		lru.Get("0")
		lru.Set(strconv.Itoa(i), i)
	}

	val, ok := lru.Get("1")
	as.False(ok)
	as.Nil(val)

	val2, ok2 := lru.Get("0")
	as.True(ok2)
	as.Equal(val2, 0)

	val3, ok3 := lru.Get("9")
	as.True(ok3)
	as.Equal(val3, 9)

	lru.Set("9", 1000)
	val4, ok4 := lru.Get("9")
	as.True(ok4)
	as.Equal(val4, 1000)

	lru.Remove("9")
	val5, ok5 := lru.Get("9")
	as.False(ok5)
	as.Nil(val5)
}
