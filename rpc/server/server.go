package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"sync"
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
	sp, e = lookupStockInDB(sr)
	return
}

func (s *stockEnquiryServer) ListNearbyStock(sr *pb.StockRequest, stream pb.StockEnquiry_ListNearbyStockServer) (e error) {
	var wg sync.WaitGroup

	// create a stock request for the 'nearest' 10 stores
	for i := 1; i <= 10; i++ {
		sri := &pb.StockRequest{
			Product: sr.Product,
			Store:   &pb.Store{StoreID: int32(i)},
		}

		// kick of our stock requests in parallel
		wg.Add(1)
		go func(sr *pb.StockRequest, stream pb.StockEnquiry_ListNearbyStockServer) {
			defer wg.Done()
			sp, _ := lookupStockInDB(sr)
			if e = stream.Send(sp); e != nil {
				grpclog.Fatalf("failed to send response: %v", e)
			}
		}(sri, stream)

	}

	// wait for our stock requests to complete
	wg.Wait()

	return nil
}

func lookupStockInDB(sr *pb.StockRequest) (sp *pb.StockPosition, e error) {

	// random wait to simulate going out to a DB or remote service
	time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)

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
