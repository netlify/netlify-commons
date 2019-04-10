package messaging

import (
	"fmt"
	"strings"

	stan "github.com/nats-io/go-nats-streaming"
	"github.com/nats-io/go-nats-streaming/pb"

	"github.com/netlify/netlify-commons/nconf"
)

func StartPoint(config *nconf.NatsConfig) (stan.SubscriptionOption, error) {
	switch v := strings.ToLower(config.StartPos); v {
	case "all":
		return stan.DeliverAllAvailable(), nil
	case "last":
		return stan.StartWithLastReceived(), nil
	case "new":
		return stan.StartAt(pb.StartPosition_NewOnly), nil
	case "", "first":
		return stan.StartAt(pb.StartPosition_First), nil
	}
	return nil, fmt.Errorf("Unknown start position '%s', possible values are all, last, new, first and ''", config.StartPos)
}
