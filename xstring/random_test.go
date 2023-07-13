package xstring

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRand(t *testing.T) {
	Seed(time.Now().UnixNano())
	assert.True(t, len(Rand()) > 0)
	assert.True(t, len(RandId()) > 0)

	const size = 10
	assert.True(t, len(RandomN(size)) == size)
}

func BenchmarkRandString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = RandomN(10)
	}
}
