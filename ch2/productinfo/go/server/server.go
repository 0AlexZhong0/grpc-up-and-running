package main

import (
	// the path the value defined for the go_package option in the .proto file
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"path/filepath"
	"strings"

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

/*
1. iterate through the items of each order in the orderMap

2. search if the item contains the query, send the order to the stream if so

3. close the stream once the iteration is finished
*/
func (s *server) SearchOrders(searchInput *orderManagementPb.SearchOrderQuery, stream orderManagementPb.OrderManagement_SearchOrdersServer) error {
	// NOTE: the method signature is that it takes the request input argument and a stream
	for orderKey, order := range s.orderMap {
		for _, orderItem := range order.Items {
			// send the current order to the consumer stream
			if strings.Contains(orderItem, searchInput.Query) {
				log.Printf("Matching Order Found: %v", orderKey)
				err := stream.Send(order)
				if err != nil {
					return fmt.Errorf("error sending the message to the stream: %v", err)
				}
				break
			}
		}
	}

	return nil
}

func (s *server) UpdateOrders(stream orderManagementPb.OrderManagement_UpdateOrdersServer) error {
	updatedOrdersStr := "Updated Order IDs: "

	for {
		order, err := stream.Recv()

		if err == io.EOF {
			return stream.SendAndClose(&wrappers.StringValue{Value: fmt.Sprintf("Orders processed %v", updatedOrdersStr)})
		}

		s.orderMap[order.Id] = order
		log.Printf("Order ID %v Updated", order.Id)
		updatedOrdersStr += order.Id + ", "
	}
}

func (s *server) ProcessOrders(stream orderManagementPb.OrderManagement_ProcessOrdersServer) error {
	batchMarker := 1
	combinedShipmentMap := make(map[string]orderManagementPb.CombinedShipment)

	for {
		orderId, err := stream.Recv()
		log.Printf("Reading Proc order: %s", orderId)
		if err == io.EOF {
			log.Printf("EOF : %s", orderId)
			// TODO: resolve the warnings
			for _, shipment := range combinedShipmentMap {
				if err := stream.Send(&shipment); err != nil {
					return err
				}
			}
			// returning nil marks as the end of the server-side stream
			return nil
		}

		if err != nil {
			return err
		}

		dest := s.orderMap[orderId.Value].Destination
		shipment, found := combinedShipmentMap[dest]

		if found {
			ord := s.orderMap[orderId.Value]
			shipment.OrdersList = append(shipment.OrdersList, ord)
			combinedShipmentMap[dest] = shipment
		} else {
			comShip := orderManagementPb.CombinedShipment{Id: "cmb - " + s.orderMap[orderId.Value].Destination, Status: "Processed"}
			ord := s.orderMap[orderId.Value]
			comShip.OrdersList = append(shipment.OrdersList, ord)
			combinedShipmentMap[dest] = comShip
			log.Println(len(comShip.OrdersList), comShip.Id)
		}

		if batchMarker == orderBatchSize {
			for _, comb := range combinedShipmentMap {
				log.Printf("Shipping : %v -> %v", comb.Id, len(comb.OrdersList))
				if err := stream.Send(&comb); err != nil {
					return err
				}
			}

			batchMarker = 0
			combinedShipmentMap = make(map[string]orderManagementPb.CombinedShipment)
		} else {
			batchMarker++
		}
	}
}

func (s *server) LoadOrders() {
	orderJsonDbPath, _ := filepath.Abs("../data/example_orders.json")
	orderData, err := ioutil.ReadFile(orderJsonDbPath)

	if err != nil {
		log.Fatalf("Failed to load default orders: %v", err)
	}

	if err := json.Unmarshal(orderData, &s.orderMap); err != nil {
		log.Fatalf("Failed to load default orders: %v", err)
	}
}

const (
	port           = ":50051"
	orderBatchSize = 3
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
