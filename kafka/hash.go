package kafka

import "hash/fnv"

func GetPartition(key string, partitions []int32, algorithm PartitionerAlgorithm) int32 {
	if len(partitions) == 0 {
		return -1
	}
	var idx uint32
	numPartitions := uint32(len(partitions))
	switch algorithm {
	case PartitionerMurMur2:
		// NOTE: the murmur2 balancers in java and librdkafka treat a nil key as
		//       non-existent while treating an empty slice as a defined value.
		idx = (murmur2([]byte(key)) & 0x7fffffff) % numPartitions
	case PartitionerFNV1A:
		idx = uint32(fnv1([]byte(key))) % numPartitions
	}
	return int32(partitions[idx])
}

// Go port of the Java library's murmur2 function.
// https://github.com/apache/kafka/blob/1.0/clients/src/main/java/org/apache/kafka/common/utils/Utils.java#L353
func murmur2(data []byte) uint32 {
	length := len(data)
	const (
		seed uint32 = 0x9747b28c
		// 'm' and 'r' are mixing constants generated offline.
		// They're not really 'magic', they just happen to work well.
		m = 0x5bd1e995
		r = 24
	)

	// Initialize the hash to a random value
	h := seed ^ uint32(length)
	length4 := length / 4

	for i := 0; i < length4; i++ {
		i4 := i * 4
		k := (uint32(data[i4+0]) & 0xff) + ((uint32(data[i4+1]) & 0xff) << 8) + ((uint32(data[i4+2]) & 0xff) << 16) + ((uint32(data[i4+3]) & 0xff) << 24)
		k *= m
		k ^= k >> r
		k *= m
		h *= m
		h ^= k
	}

	// Handle the last few bytes of the input array
	extra := length % 4
	if extra >= 3 {
		h ^= (uint32(data[(length & ^3)+2]) & 0xff) << 16
	}
	if extra >= 2 {
		h ^= (uint32(data[(length & ^3)+1]) & 0xff) << 8
	}
	if extra >= 1 {
		h ^= uint32(data[length & ^3]) & 0xff
		h *= m
	}

	h ^= h >> 13
	h *= m
	h ^= h >> 15

	return h
}

func fnv1(data []byte) uint32 {
	hasher := fnv.New32a()
	hasher.Write(data)
	return hasher.Sum32()
}
