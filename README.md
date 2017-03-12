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

To send a GOB the approach is very similar to that used to send a string.

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

Receiving a GOB is equally straight forward.

```
var data P
dec := gob.NewDecoder(rw)
err := dec.Decode(&data)
```

It isn't obvious how the decoder knows it has reached the end of the byte stream for a GOB, but looking at the `gob.Decoder` documentation it appears to be the EOF marker.

## Send the same message using protobuf

Protobuf is a little different. Rather than use a struct to hold our data, we describe our data using the proto3 format and then generate the corresponding source code. This code allows us to create instances of our protobuf and to encode them for sending over the network.

```
pbufP := &pb.P{
    X:    3,
    Y:    4,
    Z:    5,
    Name: "Pythagoras",
}

// error handling removed for brevity
out, _ := proto.Marshal(pbufP)
n, _ = rw.Write(out)
rw.Flush()
```

The `proto.Marshal()` function returns a byte slice which can then be written directly to the network. We flush the buffer as in our two previous examples and then we are done.

Receiving a protobuf message is equally straightforward, but there was one catch that tripped me up. Again, error handling has been removed from the code below for brevity.

```
pbufP := pb.P{}
d, _ := ioutil.ReadAll(rw)
proto.Unmarshal(d, &pbufP)
log.Println("server:", pbufP.String())
```

The thing that tripped me up was determining just how much data I should read from the reader. It turns out that the EOF marker is key here. When a protobuf is sent, it is terminated with the EOF marker. Using `iotil.ReadAll()` will read in all the data up to the EOF marker and hence our full protobuf.

## What about gRPC

Rather than hack this in to the existing code, I'm going to start a new client/server combination to implement this using gRPC.

There are four types of service method with gRPC:

 - simple RPC (server sends single response to client request)
 - server-side streaming RPC (server streams a sequence of messages in response to client request)
 - client-side streaming RPC (client streams events to the server and waits for a single response)
 - bidirectional streaming RPC (both server and client independently stream messages to each other)

In this proof of concept, I'm going to implement a simple Stock Enquiry Service and demonstrate the first two types of service method:

```
service StockEnquiry {
    // obtains the stock position for a given product and store
    rpc GetStockPosition(StockRequest) returns (StockPosition) {}

    // lists nearby stores that have a given item in stock
    rpc ListNearbyStock(StockRequest) returns (stream StockPosition) {}
}
```

We extend our work with protobufs by defining our service and our messages in a proto file. We then generate our service and message definitions in Go.

As with all our examples so far we begin by creating a TCP listener

```golang
l, err := net.Listen("tcp", ":10000")
```

We then create the gRPC server (no options for now) and then register our server for the Stock Enquiry service. At this point we are ready to start receiving inbound requests.

```golang
var opts []grpc.ServerOption
grpcServer := grpc.NewServer(opts...

pb.RegisterStockEnquiryServer(grpcServer, newStockEnquiryServer())
grpcServer.Serve(lis)
```

If we dig a little deeper into the definition of `RegisterStockEnquiryServer()` we can see it takes two parameters. The first is a pointer to a `grpc.Server` but the second is a `StockEnquiryServer`.

```golang
func RegisterStockEnquiryServer(s *grpc.Server, srv StockEnquiryServer) {}
```

The `StockEnquiryServer` is an interface type which includes the following method set.

```
type StockEnquiryServer interface {
    GetStockPosition(context.Context, *StockRequest) (*StockPosition, error)
    ListNearbyStock(*StockRequest, StockEnquiry_ListNearbyStockServer) error
}
```

This looks remarkably similar to the definition we outlined in our `.proto` file. Any type that implements both these methods can be registered as a `StockEnquiryServer`. This is a great way of de-coupling the generated code from code that is specific to our implementation.

Now that we know we need to pass in an instance of the `StockEnquiryServer` we need to go ahead and create one. The `newStockEnquiryServer()` function does just that.

```golang
type stockEnquiryServer struct{}

func newStockEnquiryServer() (s *stockEnquiryServer) {
    s = new(stockEnquiryServer)
    return
}

func (s *stockEnquiryServer) GetStockPosition(ctx context.Context, sr *pb.StockRequest) (sp *pb.StockPosition, e error) {
//...
}

func (s *stockEnquiryServer) ListNearbyStock(sr *pb.StockRequest, stream pb.StockEnquiry_ListNearbyStockServer) (e error) {
//...
}
```

By implementing these methods, we are able to serve gRPC requests. And, to show how simple this can be, this is our `GetStockPosition()` method.

```golang
func (s *stockEnquiryServer) GetStockPosition(ctx context.Context, sr *pb.StockRequest) (sp *pb.StockPosition, e error) {
    sp, e = lookupStockInDB(sr)
    return
}
```

## Notes:

 - Protobuf RPC definitions only allow for single parameters. See [#976](https://github.com/google/protobuf/issues/976) on GitHub
 - Understanding how the `Read()` method is implemented on a TCP connection quickly drops down into the `syscall` package.
