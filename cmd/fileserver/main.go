package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/dledbetter123/distributed-prime-counter/pkg/config"
	primes "github.com/dledbetter123/distributed-prime-counter/pkg/primespb"
)

type fileServer struct {
	primes.UnimplementedFileServerServer
	dataFilePath string
	chunkSize    int
}

func (s *fileServer) Shutdown(ctx context.Context, req *primes.ShutdownRequest) (*primes.Ack, error) {
	log.Println("Shutting down the file server.")
	os.Exit(0)
	return &primes.Ack{Success: true}, nil
}

func (s *fileServer) GetFileChunk(req *primes.JobRequest, stream primes.FileServer_GetFileChunkServer) error {
	file, err := os.Open(req.Job.Pathname)
	if err != nil {
		return err
	}
	defer file.Close()

	// start position of the job
	_, err = file.Seek(req.Job.Start, 0)
	if err != nil {
		return err
	}

	buffer := make([]byte, s.chunkSize)
	remaining := int64(req.Job.Length)
	for remaining > 0 {
		readSize := int64(s.chunkSize)
		if int64(readSize) > remaining {
			readSize = int64(remaining)
		}
		n, err := file.Read(buffer[:readSize])
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if err := stream.Send(&primes.FileChunk{ChunkData: buffer[:n]}); err != nil {
			return err
		}
		remaining -= int64(n)
	}
	return nil
}

func main() {

	var configPath string
	flag.StringVar(&configPath, "config", "../primes_config.txt", "Path to configuration file")
	chunkSize := flag.Int("C", 1024, "Chunk size in bytes")
	dataFilePath := flag.String("datafile", "../data/newgen.dat", "Path to the data file")
	flag.Parse()
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Printf("Starting fileserver, managing datafile: %s\n", *dataFilePath)

	startFileServer(cfg.FileServer, *chunkSize, *dataFilePath)
}

func startFileServer(address string, chunkSize int, filepath string) {
	fmt.Printf("Starting fileserver, listening on %s\n", address)

	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	primes.RegisterFileServerServer(s, &fileServer{
		dataFilePath: filepath,
		chunkSize:    chunkSize,
	})
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
