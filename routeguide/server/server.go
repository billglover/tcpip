package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	pb "github.com/billglover/tcpip/routeguide/routeguide"
)

var (
	port = flag.Int("port", 10000, "The server port")
)

type routeGuideServer struct {
	savedFeatures []*pb.Feature
}

// GetFeature returns the feature at the given point.
func (s *routeGuideServer) GetFeature(ctx context.Context, point *pb.Point) (*pb.Feature, error) {
	for _, feature := range s.savedFeatures {
		if proto.Equal(feature.Location, point) {
			return feature, nil
		}
	}
	// No feature was found, return an unnamed feature
	return &pb.Feature{Location: point}, nil
}

func newServer() *routeGuideServer {
	s := new(routeGuideServer)
	p := pb.Point{Latitude: 1, Longitude: 1}
	f := pb.Feature{Name: "feature_1", Location: &p}
	s.savedFeatures = make([]*pb.Feature, 1)
	s.savedFeatures[0] = &f
	return s
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		grpclog.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterRouteGuideServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
