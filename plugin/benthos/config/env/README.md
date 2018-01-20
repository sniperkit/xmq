Environment Vars Config
=======================

The default environment variables based config is a simple bridge config with
env variables for all important fields.

Values for `BENTHOS_INPUT` and `BENTHOS_OUTPUT` should be chosen from [here][0]
and [here][1] respectively.

The full list of variables and their default values:

``` sh
BENTHOS_HTTP_ADDRESS           = 0.0.0.0:4195
BENTHOS_INPUT                  =
BENTHOS_OUTPUT                 =
BENTHOS_LOG_LEVEL              = INFO

BENTHOS_MAX_PARTS              = 100
BENTHOS_MIN_PARTS              = 1
BENTHOS_MAX_PART_SIZE          = 1073741824
BENTHOS_INPUT_PROCESSOR        = noop
BENTHOS_OUTPUT_PROCESSOR       = noop

ZMQ_INPUT_URLS                 =
ZMQ_INPUT_BIND                 = true
ZMQ_INPUT_SOCKET               = PULL
ZMQ_OUTPUT_URLS                =
ZMQ_OUTPUT_BIND                = true
ZMQ_OUTPUT_SOCKET              = PULL

SCALE_PROTO_INPUT_URLS         =
SCALE_PROTO_INPUT_BIND         = true
SCALE_PROTO_INPUT_SOCKET       = PULL
SCALE_PROTO_OUTPUT_URLS        =
SCALE_PROTO_OUTPUT_BIND        = true
SCALE_PROTO_OUTPUT_SOCKET      = PULL

KAFKA_INPUT_BROKER_ADDRESSES   =
KAFKA_INPUT_CLIENT_ID          = benthos-client
KAFKA_INPUT_CONSUMER_GROUP     = benthos-group
KAFKA_INPUT_TOPIC              = benthos-stream
KAFKA_INPUT_PARTITION          = 0
KAFKA_INPUT_START_OLDEST       = true
KAFKA_OUTPUT_BROKER_ADDRESSES  =
KAFKA_OUTPUT_CLIENT_ID         = benthos-client
KAFKA_OUTPUT_CONSUMER_GROUP    = benthos-group
KAFKA_OUTPUT_TOPIC             = benthos-stream
KAFKA_OUTPUT_PARTITION         = 0
KAFKA_OUTPUT_START_OLDEST      = true
KAFKA_OUTPUT_ACK_REP           = true

AMQP_INPUT_URL                 =
AMQP_INPUT_EXCHANGE            = benthos-exchange
AMQP_INPUT_EXCHANGE_TYPE       = direct
AMQP_INPUT_QUEUE               = benthos-stream
AMQP_INPUT_KEY                 = benthos-key
AMQP_INPUT_CONSUMER_TAG        = benthos-consumer
AMQP_OUTPUT_URL                =
AMQP_OUTPUT_EXCHANGE           = benthos-exchange
AMQP_OUTPUT_EXCHANGE_TYPE      = direct
AMQP_OUTPUT_QUEUE              = benthos-stream
AMQP_OUTPUT_KEY                = benthos-key
AMQP_OUTPUT_CONSUMER_TAG       = benthos-consumer

NSQD_INPUT_TCP_ADDRESSES       =
NSQD_INPUT_LOOKUP_ADDRESSES    =
NSQ_INPUT_TOPIC                = benthos-messages
NSQ_INPUT_CHANNEL              = benthos-stream
NSQ_INPUT_USER_AGENT           = benthos-consumer
NSQ_OUTPUT_TCP_ADDRESS         =
NSQ_OUTPUT_TOPIC               = benthos-messages
NSQ_OUTPUT_CHANNEL             = benthos-stream
NSQ_OUTPUT_USER_AGENT          = benthos-consumer

NATS_INPUT_URLS                =
NATS_INPUT_SUBJECT             = benthos-stream
NATS_INPUT_CLUSTER_ID          = benthos-cluster  # Used only for nats_stream
NATS_INPUT_CLIENT_ID           = benthos-consumer # ^
NATS_INPUT_QUEUE               = benthos-queue    # ^
NATS_INPUT_DURABLE_NAME        = benthos-offset   # ^
NATS_OUTPUT_URLS               =
NATS_OUTPUT_SUBJECT            = benthos-stream
NATS_OUTPUT_CLUSTER_ID         = benthos-cluster  # Used only for nats_stream
NATS_OUTPUT_CLIENT_ID          = benthos-consumer # ^

REDIS_INPUT_URL                =
REDIS_INPUT_CHANNEL            = benthos-stream
REDIS_OUTPUT_URL               =
REDIS_OUTPUT_CHANNEL           = benthos-stream

HTTP_SERVER_INPUT_ADDRESS      =
HTTP_SERVER_INPUT_PATH         = /post
HTTP_SERVER_OUTPUT_ADDRESS     =
HTTP_SERVER_OUTPUT_PATH        = /get
HTTP_SERVER_OUTPUT_STREAM_PATH = /get/stream

HTTP_CLIENT_INPUT_URL          =
HTTP_CLIENT_OUTPUT_URL         =

FILE_INPUT_PATH                =
FILE_INPUT_MULTIPART           = false
FILE_INPUT_MAX_BUFFER          = 65536
FILE_OUTPUT_PATH               =
FILE_OUTPUT_MULTIPART          = false
FILE_OUTPUT_MAX_BUFFER         = 65536
```

[0]: ../../resources/docs/inputs/list.md
[1]: ../../resources/docs/outputs/list.md
