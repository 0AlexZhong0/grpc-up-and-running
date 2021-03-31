package main

import (
	"context"
	"io"
	"log"
	"time"

	orderManagementPb "github.com/0AlexZhong0/grpc-up-and-running-protos/order_management"
	pb "github.com/0AlexZhong0/grpc-up-and-running-protos/productinfo"
	"github.com/golang/protobuf/ptypes/wrappers"

	"google.golang.org/grpc"
)

const (
	address = "localhost:50051"
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewProductInfoClient(conn)
	orderManagementClient := orderManagementPb.NewOrderManagementClient(conn)

	name := "Apple iPhone 11"
	description := "Meet Apple iPhone 11.All-new dual-camera system with Ultra Wide and Night mode."

	price := float32(1000.0)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.AddProduct(ctx, &pb.Product{Name: name, Description: description, Price: price})

	if err != nil {
		log.Fatalf("Could not add product: %v", err)
	}

	log.Printf("Product ID: %s added successfully", r.Value)

	product, err := c.GetProduct(ctx, &pb.ProductID{Value: r.Value})
	if err != nil {
		log.Fatalf("Could not get product: %v", err)
	}

	log.Printf("Product: %v", product.String())

	// order management
	searchStream, _ := orderManagementClient.SearchOrders(ctx, &orderManagementPb.SearchOrderQuery{Query: "Si"})
	for {
		searchOrder, err := searchStream.Recv()

		if err == io.EOF {
			break
		}

		log.Print("Search Result: ", searchOrder)
	}

	// Process Orders: Bi-directional Streamings
	streamProcOrder, err := orderManagementClient.ProcessOrders(ctx)
	if err != nil {
		log.Fatalf("%v.ProcessOrders(_) = _, %v", orderManagementClient, err)
	}

	if err := streamProcOrder.Send(&wrappers.StringValue{Value: "1"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", orderManagementClient, "1", err)
	}

	// use channel and goroutine to read and write on two threads concurrently
	// channel := make(chan struct{})

	// go asyncClientBidirectionalRPC(streamProcOrder, channel)
	// time.Sleep(time.Second)

	if err := streamProcOrder.Send(&wrappers.StringValue{Value: "2"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", orderManagementClient, "2", err)
	}

	if err := streamProcOrder.Send(&wrappers.StringValue{Value: "3"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", orderManagementClient, "3", err)
	}

	// <-channel
}

func asyncClientBidirectionalRPC(streamProcOrder orderManagementPb.OrderManagement_ProcessOrdersClient, c chan struct{}) {
	for {
		combinedShipment, errProcOrder := streamProcOrder.Recv()
		if errProcOrder == io.EOF {
			break
		}

		log.Println("Combined shipment : ", combinedShipment.OrdersList)
	}

	<-c
}
