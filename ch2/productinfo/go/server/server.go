package main

import (
	// the path the value defined for the go_package option in the .proto file
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"path/filepath"

	orderManagementPb "github.com/0AlexZhong0/grpc-up-and-running-protos/order_management"
	pb "github.com/0AlexZhong0/grpc-up-and-running-protos/productinfo"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	pb.UnimplementedProductInfoServer
	orderManagementPb.UnimplementedOrderManagementServer

	productMap map[string]*pb.Product
	orderMap   map[string]*orderManagementPb.Order
}

func (s *server) AddProduct(ctx context.Context, in *pb.Product) (*pb.ProductID, error) {
	out, err := uuid.NewUUID()

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error while generating Product ID", err)
	}

	in.Id = out.String()
	if s.productMap == nil {
		s.productMap = make(map[string]*pb.Product)
	}

	s.productMap[in.Id] = in
	return &pb.ProductID{Value: in.Id}, status.New(codes.OK, "").Err()
}

func (s *server) GetProduct(ctx context.Context, in *pb.ProductID) (*pb.Product, error) {
	value, exists := s.productMap[in.Value]
	if exists {
		return value, status.New(codes.OK, "").Err()
	}

	return nil, status.Errorf(codes.NotFound, "Product does not exist", in.Value)
}

// order management service methods
func (s *server) GetOrder(ctx context.Context, in *wrappers.StringValue) (*orderManagementPb.Order, error) {
	resultOrder, exists := s.orderMap[in.Value]

	if !exists {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Order %v is not found", in.Value))
	}

	return resultOrder, status.New(codes.OK, "").Err()
}

func (s *server) LoadOrders() {
	orderJsonDbPath, _ := filepath.Abs("./data/example_orders.json")
	orderData, err := ioutil.ReadFile(orderJsonDbPath)

	if err != nil {
		log.Fatalf("Failed to load default orders: %v", err)
	}

	if err := json.Unmarshal(orderData, &s.orderMap); err != nil {
		log.Fatalf("Failed to load default orders: %v", err)
	}
}

const (
	port = ":50051"
)

func newServer() *server {
	s := &server{}
	s.LoadOrders()
	return s
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	productServer := newServer()
	pb.RegisterProductInfoServer(s, productServer)
	orderManagementPb.RegisterOrderManagementServer(s, productServer)

	log.Printf("Starting gRPC listener on port " + port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
