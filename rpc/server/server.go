package main

import (
	"flag"
	"fmt"
	"net"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	pb "github.com/billglover/tcpip/rpc/stockenq"
)

var (
	port = flag.Int("port", 10000, "The server port")
)

type stockEnquiryServer struct{}

func newStockEnquiryServer() (s *stockEnquiryServer) {
	s = new(stockEnquiryServer)
	return
}

func (s *stockEnquiryServer) GetStockPosition(ctx context.Context, sr *pb.StockRequest) (sp *pb.StockPosition, e error) {

	// return a dummy stock position
	sp = &pb.StockPosition{
		Product:       sr.Product,
		Store:         sr.Store,
		Units:         5,
		NextAvailable: time.Now().UnixNano(),
		Status:        pb.StockPosition_IN_STOCK,
	}

	return
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		grpclog.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterStockEnquiryServer(grpcServer, newStockEnquiryServer())
	grpcServer.Serve(lis)
}
