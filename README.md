# tKeel amqp
We provide `sender` and `receiver` for you to easily connect **amqp** using the Go.
## Required
Please check your environment have the following lib:

[Qpid Proton - AMQP messaging toolkit](https://github.com/apache/qpid-proton/blob/main/INSTALL.md)


## Usage
Default Start the server at `:3172`

Start the server:
```bash
go run ./cmd/* --address :3172
```

More details please see the `main.go` file for how this server running.

## Receiver Usage
See `receiver_test.go`

That's very simple.
```go
testUrl := "amqp://localhost:3172/topic"
r, err := NewReceiver(testUrl, electron.User("fred1"),
    electron.VirtualHost("token"),
    electron.Password([]byte("mypassword")),
    electron.SASLAllowInsecure(true))
assert.NoError(t, err)
for i := 0; i < 5; i++ {
    content, err := r.Receive()
    assert.NoError(t, err)
    fmt.Printf("Received %d: %s\n", i, content)
}
```

## Build and start tKeel amqp-broker Server
Build amqp-broker server:
```bash
go build -o ./bin/amqp-broker ./cmd/*.go
```
Then binary executable file is `amqp-broker` in `./bin`.

Config your system environment variable switch your kafka connection config info: 

(Source code Detail see: `internal/mq/config.go`)
```bash
    export KAFKA_BROKERS=localhost:3172
    export KAFKA_CONSUMER_GROUP=amqp
    export KAFKA_VERSION=3.1.0
    export KAFKA_ASSIGNOR=roundrobin
    export KAFKA_OLDEST_ENABLE=true
```

Config `Core-Broker` connection info:
```bash
    export CORE_BROKER_SUBSCRIBE_VALIDATE_URL=http://192.168.123.9:30707/apis/core-broker/v1/validate/subscribe
```

## Build amqp receiver tool
```bash
go build -o ./bin/amqp-receiver ./cmd/reciver/*.go
```

Run it:
```bash
./bin/amqp-receiver -c amqp://tkeel.io:30082
```

More flag or helping see `-h` flag
```bash
./bin/amqp-receiver -h
```