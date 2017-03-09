# tcpip
A look at TCP/IP networking in Go

Based on this blog post by Christoph Berger, [TCP/IP Networking](https://appliedgo.net/networking/)

## Goals

 - Basic: send and receive a simple message as a string
 - Basic: send and receive a struct via GOB
 - Stretch: send the same messages using Protobuf
 - Stretch: what about grpc

## Send and receive a simple message as a string

```
listener, e := net.Listen("tcp", port)
```

This creates a listener for stream-oriented network protocols (in this case TCP). The listener is a simple interface.

```
type Listener interface {
        Accept() (Conn, error)
        Close() error
        Addr() Addr
}
```

 > How does `listener` implement the `Accept()` method?

The accept method returns a `Conn` which is itself an interface that specifies the following methods/

```
type Conn interface {
        Read(b []byte) (n int, err error)
        Write(b []byte) (n int, err error)
        Close() error
        LocalAddr() Addr
        RemoteAddr() Addr
        SetDeadline(t time.Time) error
        SetReadDeadline(t time.Time) error
        SetWriteDeadline(t time.Time) error
}
```

The implementations of both the `Listener` and the `Conn` interfaces are platform specific.

## Send and receive a struct via GOB

Much of the code to send a GOB is the same as that to send a string. To send a string, we call the `WriteString()` method on the ReadWriter.

```
n, err = rw.WriteString("Additional data.\n")
```

To send a GOB the approach is very similar.

```
enc := gob.NewEncoder(rw)
err = enc.Encode(P{3, 4, 5, "Pythagoras"})
```

The `NewEncoder(w io.Writer)` method takes an io.Writer and returns an instance of the `Encoder` struct. Calling `Encode(e interface{})` writes the encoded version of the interface out as a byte stream.

The GOB encoder handles recursive types as in the example below.

```
type P struct {
    X, Y, Z int
    Name    string
    P       *P
}
```

## Send the same message using protobuf