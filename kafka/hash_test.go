package kafka

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMurMurHash2(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]int32{
		"kafka":                                4,
		"35235629-3f4a-4243-bac9-cc0d2d5014b1": 6,
		"162a7f77-fea4-4f22-9746-6d635a93bc57": 3,
		"2b6172c5-bddd-4746-bc5c-d1f9316a85ca": 4,
		"dfd1b4ef-a06e-4380-8952-182e0a7718bf": 4,
		"b27f5f66-b380-4ba1-add3-8af636fe9620": 1,
		"":                                     1,
	}

	partitions := []int32{0, 1, 2, 3, 4, 5, 6, 7}

	for key, part := range tests {
		p := GetPartition(key, partitions, PartitionerMurMur2)
		assert.Equal(part, p)
	}
}

func TestFNV1Hash(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]int32{
		"kafka": 2,
		"b3913c28935540691f7758cd25b58c97ca8137a5bf4a61a95be45f378e8e9982": 2,
		"35235629-3f4a-4243-bac9-cc0d2d5014b1":                             2,
		"162a7f77-fea4-4f22-9746-6d635a93bc57":                             1,
		"2b6172c5-bddd-4746-bc5c-d1f9316a85ca":                             2,
		"dfd1b4ef-a06e-4380-8952-182e0a7718bf":                             0,
		"b27f5f66-b380-4ba1-add3-8af636fe9620":                             1,
		"":                                                                 2,
	}

	partitions := []int32{0, 1, 2}

	for key, part := range tests {
		p := GetPartition(key, partitions, PartitionerFNV1A)
		assert.Equal(part, p, key)
	}
}

func TestFilebeatHash(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]int32{
		"kafka": 0,
		"b3913c28935540691f7758cd25b58c97ca8137a5bf4a61a95be45f378e8e9982": 0,
		"35235629-3f4a-4243-bac9-cc0d2d5014b1":                             0,
		"162a7f77-fea4-4f22-9746-6d635a93bc57":                             1,
		"2b6172c5-bddd-4746-bc5c-d1f9316a85ca":                             2,
		"dfd1b4ef-a06e-4380-8952-182e0a7718bf":                             1,
		"b27f5f66-b380-4ba1-add3-8af636fe9620":                             1,
		"":                                                                 1,
	}

	partitions := []int32{0, 1, 2}

	for key, part := range tests {
		p := GetPartition(key, partitions, PartitionerFilebeat)
		assert.Equal(part, p, key)
	}
}
