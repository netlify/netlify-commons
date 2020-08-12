# github.com/netlify/netlify-commons/kafka

Package kafka provides a Consumer and a Producer for basic Kafka operations.

It relies on https://github.com/confluentinc/confluent-kafka-go which is a Go wrapper on top of https://github.com/edenhill/librdkafka.
This provides a reliable implementation, fully supported by the community, but also from Confluent, the creators of Kafka.

## Docs

Please find the generated **godoc** documentation including some examples in [pkg.go.dev](https://pkg.go.dev/mod/github.com/netlify/netlify-commons?tab=packages).

## TODO

- Support standalone consumers and not only consumers members of a consumer group. 
- Support seeking by timestamp (only offset is supported)
- Integration tests