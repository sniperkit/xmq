Benthos Docker
==============

This is a multi stage Dockerfile that builds Benthos and then copies it to a
scratch image. The image comes with a config that allows you to configure simple
bridges using [environment variables](../../config/env/README.md) like this:

``` sh
docker run \
	-e "BENTHOS_INPUT=kafka_balanced" \
	-e "KAFKA_INPUT_BROKER_ADDRESSES=foo:6379" \
	-e "BENTHOS_OUTPUT=nats" \
	-e "NATS_OUTPUT_URLS=nats://bar:4222,nats://baz:4222" \
	benthos
```

Alternatively, you can run the image using a custom config file:

``` sh
docker run --rm -v /path/to/your/config.yaml:/benthos.yaml benthos
```
