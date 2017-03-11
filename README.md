# tcpip
A look at TCP/IP networking in Go

Based on this blog post by Christoph Berger, [TCP/IP Networking](https://appliedgo.net/networking/)

## Goals

 - Basic: establishing network connectivity
 - Basic: send and receive a simple message as a string
 - Basic: send and receive a struct via GOB
 - Stretch: send the same messages using Protobuf
 - Stretch: what about grpc

## Establishing network connectivity

Our application operates in two different configurations; client and server. In server mode, we listen for an inbound TCP connection on a specified port. It is as simple as this.

```
listener, e := net.Listen("tcp", port)
```

This creates a listener for stream-oriented network protocols (in this case TCP). The listener is an interface.

```
type Listener interface {
        Accept() (Conn, error)
        Close() error
        Addr() Addr
}
```

The accept method returns a `Conn` which is itself an interface that has the following methods.

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

The `Read(b []byte) (n int, err error)` method means that the `Conn` conforms to the `io.Reader` interface. The `Write(b []byte) (n int, err error)` method means that the `Conn` fonforms to the `io.Writer` interface.

This allows us to create a new `bufio.ReadWriter` with the following command.

```
rw := bufio.NewReadWriter(bufio.NewReader(c), bufio.NewWriter(c))
```

Our server is now all set for reading and writing bytes over our TCP connection.

The client itself is a little simpler. Instead of listening for an incomming connection, we need to `net.Dial()` a server. This returns a `Conn` which we have already seen can be used to establish a `bufio.ReadWriter` for reading and writing bytes over the connection.

```
conn, err := net.Dial("tcp", addr)
// error handling omitted for brevity
rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
```

With both our client and our server able to read and write bytes, we are now ready to look at different data structures.

## Send and receive a simple message as a string

Writing a string to a `bufio.ReadWriter` is as simple as doing the following.

```
n, err := rw.WriteString("STRING\n")
```

This writes a string and returns the number of bytes written. If the number of bytes written is less than the length of the string, it also returns an error.

There is only one final thing we need to do in order to send our string over the network, and that is to flush the buffered data to the underlying writer.

```
err = rw.Flush()
```

And with that we are done, our string has been sent over the network where hopefully our server has received it.

Reading a string is fairly straight forward.

```
cmd, err := rw.ReadString('\n')
```

Note though that we are providing a string terminator, in this case the new line charcter, `\n`, as a means of determining when we have received a complete string. This ability to know when we are done receiving data is crucial as without it, the server has no way of knowing if the message it should be receiving is `ST`, `STRING`, or `STRING MESSAGE`.

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