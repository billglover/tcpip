package main

import (
	"context"
	"flag"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	pb "github.com/billglover/tcpip/rpc/stockenq"
)

var (
	serverAddr = flag.String("server_addr", "127.0.0.1:10000", "The server address in the format of host:port")
)

func getStockPosition(client pb.StockEnquiryClient, sr *pb.StockRequest) {
	grpclog.Printf("getting stock position for product '%d' in store '%d'", sr.Product.ProductCode, sr.Store.StoreID)
	sp, err := client.GetStockPosition(context.Background(), sr)
	if err != nil {
		grpclog.Fatalf("%v.GetStockPosition(_) = _, %v: ", client, err)
	}

	grpclog.Printf("%v %v\n", sp.GetStatus(), sp)
}

func getNearbyStock(client pb.StockEnquiryClient, sr *pb.StockRequest) {
	grpclog.Printf("getting nearby stock for product '%d' in store '%d'", sr.Product.ProductCode, sr.Store.StoreID)
	stream, err := client.ListNearbyStock(context.Background(), sr)
	if err != nil {
		grpclog.Fatalf("%v.ListNearbyStock(_) = _, %v: ", client, err)
	}

	for {
		sp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			grpclog.Fatalf("%v.ListNearbyStock(_) = _, %v: ", client, err)
		}
		grpclog.Printf("%v %v\n", sp.GetStatus(), sp)
	}
}

func main() {
	flag.Parse()
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		grpclog.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	// make a mock stock request
	client := pb.NewStockEnquiryClient(conn)

	p1 := pb.Product{ProductCode: 1}
	s1 := pb.Store{StoreID: 1}
	sr1 := pb.StockRequest{Product: &p1, Store: &s1}

	getStockPosition(client, &sr1)

	getNearbyStock(client, &sr1)
}
