# tKeel amqp
We provide `sender` and `receiver` for you to easily connect **amqp** using the Go.
## Required
Please check your environment have the following lib:

[Qpid Proton - AMQP messaging toolkit](https://github.com/apache/qpid-proton/blob/main/INSTALL.md)


## Usage
Default Start the server at `:3172`

Start the server:
```bash
go run ./...
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
