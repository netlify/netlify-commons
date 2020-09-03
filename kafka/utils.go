package kafka

import (
	"fmt"

	kafkalib "github.com/confluentinc/confluent-kafka-go/kafka"
)

// getPartitionIds returns the partition IDs for a given topic
func getPartitionIds(topic string, meta *kafkalib.Metadata) ([]int32, error) {
	topicMeta, ok := meta.Topics[topic]
	if !ok {
		return nil, fmt.Errorf("no metadata for given topic: %s", topic)
	}

	if topicMeta.Error.Code() != 0 {
		return nil, fmt.Errorf("%s", topicMeta.Error.Error())
	}

	partitions := topicMeta.Partitions
	idxs := make([]int32, len(partitions))

	for idx, part := range partitions {
		idxs[idx] = part.ID
	}

	return idxs, nil
}
