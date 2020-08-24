package kafka

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMurMurHash2(t *testing.T) {
	assert := assert.New(t)

	keysToTest := []string{
		"kafka",
		"35235629-3f4a-4243-bac9-cc0d2d5014b1",
		"162a7f77-fea4-4f22-9746-6d635a93bc57",
		"2b6172c5-bddd-4746-bc5c-d1f9316a85ca",
		"dfd1b4ef-a06e-4380-8952-182e0a7718bf",
		"b27f5f66-b380-4ba1-add3-8af636fe9620",
		"",
	}

	hashes := []int32{
		4, // kafka
		6, // 35235629-3f4a-4243-bac9-cc0d2d5014b1"
		3, // 162a7f77-fea4-4f22-9746-6d635a93bc57"
		4, // 2b6172c5-bddd-4746-bc5c-d1f9316a85ca"
		4, // dfd1b4ef-a06e-4380-8952-182e0a7718bf"
		1, // b27f5f66-b380-4ba1-add3-8af636fe9620"
		1, // ""
	}

	partitions := []int32{0, 1, 2, 3, 4, 5, 6, 7}

	for idx, key := range keysToTest {
		h := GetPartition(key, partitions)
		assert.Equal(hashes[idx], h)
	}
}
