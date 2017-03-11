package main

import (
	"context"
	"flag"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	pb "github.com/billglover/tcpip/routeguide/routeguide"
)

var (
	serverAddr = flag.String("server_addr", "127.0.0.1:10000", "The server address in the format of host:port")
)

// printFeature gets the feature for the given point.
func printFeature(client pb.RouteGuideClient, point *pb.Point) {
	grpclog.Printf("Getting feature for point (%d, %d)", point.Latitude, point.Longitude)
	feature, err := client.GetFeature(context.Background(), point)
	if err != nil {
		grpclog.Fatalf("%v.GetFeatures(_) = _, %v: ", client, err)
	}
	grpclog.Println(feature)
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
	client := pb.NewRouteGuideClient(conn)

	// Looking for a valid feature
	printFeature(client, &pb.Point{Latitude: 1, Longitude: 1})

}
