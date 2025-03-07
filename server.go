package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	parser "hparserGO/proto/go"
)

var db *sql.DB

func initDB() error {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	host := os.Getenv("DB_HOST")
	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		log.Fatalf("Invalid DB_PORT: %v", err)
	}
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var dbErr error
	db, dbErr = sql.Open("postgres", connectionString)
	if dbErr != nil {
		return dbErr
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	log.Println("Connected to the database")
	return nil
}

type parserServiceServer struct {
	parser.UnimplementedParserServiceServer
}

func (s *parserServiceServer) UpdateChannels(ctx context.Context, req *parser.UpdateChannelsRequest) (*parser.Channels, error) {
	log.Printf("UpdateChannels called with request: %+v", req)

	insertStmt := `
        INSERT INTO channels("update", "process_time") 
        VALUES($1, $2) RETURNING id, name;
    `
	var channelID, channelName string
	processTime := time.Now()

	err := db.QueryRow(insertStmt, req.Update, processTime).Scan(&channelID, &channelName)
	if err != nil {
		log.Printf("Error updating channels: %v", err)
		return nil, err
	}

	return &parser.Channels{
		Ids:         channelID,
		Names:       channelName,
		ProcessTime: timestamppb.New(processTime),
	}, nil
}

func (s *parserServiceServer) GetChannels(ctx context.Context, req *parser.GetChannelsRequest) (*parser.Channels, error) {
	log.Printf("GetChannels called with request: %+v", req)

	query := `
        SELECT id, url, CURRENT_TIMESTAMP AS process_time 
        FROM channels;
    `

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Error querying channels: %v", err)
		return nil, err
	}
	defer rows.Close()

	var allChannels []string
	var allNames []string
	var allProcessTimes []*timestamppb.Timestamp

	for rows.Next() {
		var channelID, channelName string
		var processTime time.Time

		err := rows.Scan(&channelID, &channelName, &processTime)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}

		allChannels = append(allChannels, channelID)
		allNames = append(allNames, channelName)
		allProcessTimes = append(allProcessTimes, timestamppb.New(processTime))
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
		return nil, err
	}

	combinedIDs := strings.Join(allChannels, ",")
	combinedUrls := strings.Join(allNames, ", ")

	if len(allProcessTimes) > 0 {
		return &parser.Channels{
			Ids:         combinedIDs,
			Names:       combinedUrls,
			ProcessTime: allProcessTimes[len(allProcessTimes)-1], // Use the last process_time
		}, nil
	}

	return &parser.Channels{}, nil
}

func main() {
	err := initDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	parser.RegisterParserServiceServer(grpcServer, &parserServiceServer{})

	log.Println("Starting gRPC server on :9090...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
