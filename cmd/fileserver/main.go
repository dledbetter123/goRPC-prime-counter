package main

import (
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

type server struct {
	primes.UnimplementedFileServerServer
	dataFilePath string
	chunkSize    int
}

func (s *server) GetFileChunk(req *primes.JobRequest, stream primes.FileServer_GetFileChunkServer) error {
	file, err := os.Open(req.Job.Pathname)
	if err != nil {
		return err
	}
	defer file.Close()

	// Seek to the start position of the job
	_, err = file.Seek(req.Job.Start, 0)
	if err != nil {
		return err
	}

	// Read the specified length from the file
	buffer := make([]byte, req.Job.Length)
	readLength, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return err
	}
	buffer = buffer[:readLength]

	// Stream the buffer
	return stream.Send(&primes.FileChunk{ChunkData: buffer})
}

func main() {
	flag.Parse()

	var configPath string
	flag.StringVar(&configPath, "config", "../primes_config.txt", "Path to configuration file")
	chunkSize := flag.Int("chunk", 1024, "Chunk size in bytes")
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
	primes.RegisterFileServerServer(s, &server{
		dataFilePath: filepath,
		chunkSize:    chunkSize,
	})
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
