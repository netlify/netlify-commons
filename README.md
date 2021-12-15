# netlify-commons

[![](https://godoc.org/github.com/netlify/netlify-commons?status.svg)](https://godoc.org/github.com/netlify/netlify-commons)

This is a core library that will add common features for our services.

> The ones that have their own right now will be migrated as needed

Mostly this deals with configuring logging, messaging (rabbit && nats), and loading configuration.

## Testing
### Prerequisites

If running on Apple silicon, the `librdkafka` library will need to be linked dynamically. We may want to keep an eye on [issues](https://github.com/confluentinc/confluent-kafka-go/issues/696) in the `confluent-kakfa-go` repository for alternative approaches that we could use in the future.

```
brew install openssl
brew install librdkafka
brew install pkg-config
```

And add `PKG_CONFIG_PATH` to your `~/.bashrc` or `~/.zshrc` (as instructed by `brew info openssl`)
```
export PKG_CONFIG_PATH="/opt/homebrew/opt/openssl@3/lib/pkgconfig"
```
