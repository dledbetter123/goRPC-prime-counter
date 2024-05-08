package main

import (
	"context"
	"flag"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/dledbetter123/distributed-prime-counter/pkg/config"
	primes "github.com/dledbetter123/distributed-prime-counter/pkg/primespb"
)

// Consolidator service
type consolidatorServer struct {
	primes.UnimplementedConsolidatorServer
	totalPrimes int32
	mu          sync.Mutex
}

func (s *consolidatorServer) SendResult(ctx context.Context, res *primes.Result) (*primes.Ack, error) {
	s.mu.Lock()
	s.totalPrimes += res.Count
	s.mu.Unlock()
	log.Printf("Received result: %v, Total primes: %d", res, s.totalPrimes)
	return &primes.Ack{Success: true}, nil
}

func startServer(srv *grpc.Server, port string) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	reflection.Register(srv)
	log.Printf("Consolidator server listening on %s", port)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func main() {
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig("../primes_config.txt")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Println("Starting the consolidator service...")
	consolidatorSrv := grpc.NewServer()
	primes.RegisterConsolidatorServer(consolidatorSrv, &consolidatorServer{})
	startServer(consolidatorSrv, cfg.Consolidator)
}
