package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	_ "google.golang.org/protobuf/types/known/timestamppb"
	parser "hparserGO/proto/go"
	"log"
	"time"
)

func updateChannels(client parser.ParserServiceClient, update string) {
	log.Printf("Updating channels with update: '%s'", update)

	response, err := client.UpdateChannels(context.Background(), &parser.UpdateChannelsRequest{
		Update: update,
	})

	if err != nil {
		log.Fatalf("Error updating channels: %v", err)
	}

	log.Printf("Updated Channels: { Ids: '%s', Names: '%s', ProcessTime: '%s' }",
		response.Ids, response.Names, response.ProcessTime.AsTime().Format(time.RFC3339))
}

func getChannels(client parser.ParserServiceClient, names string) {
	log.Printf("Getting channels with names: '%s'", names)

	response, err := client.GetChannels(context.Background(), &parser.GetChannelsRequest{
		Names: names,
	})

	if err != nil {
		log.Fatalf("Error getting channels: %v", err)
	}

	log.Printf("Retrieved Channels: { Ids: '%s', Names: '%s', ProcessTime: '%s' }",
		response.Ids, response.Names, response.ProcessTime.AsTime().Format(time.RFC3339))
}

func main() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.Dial("localhost:9090", opts...)
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	client := parser.NewParserServiceClient(conn)

	//log.Println("==== Calling UpdateChannels ====")
	//updateChannels(client, "NewChannelUpdate1")

	log.Println("==== Calling GetChannels ====")
	getChannels(client, "ExistingChannel")
}
