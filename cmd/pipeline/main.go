package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/dledbetter123/distributed-prime-counter/pkg/config"
	primes "github.com/dledbetter123/distributed-prime-counter/pkg/primespb"
)

var (
	jobs         chan primes.Job
	dataFilePath = flag.String("datafile", "../data/newgen.dat", "Path to the binary data file")
	N            = flag.Int("N", 65536, "Size of each file segment in KB")
)

type dispatcherServer struct {
	primes.UnimplementedDispatcherServer
	totalJobs          int32
	completedJobs      int32
	mu                 sync.Mutex
	startTime          time.Time // Variable to store the start time
	consolidatorClient primes.ConsolidatorClient
	fileServerClient   primes.FileServerClient
}

func NewDispatcherServer() *dispatcherServer {
	return &dispatcherServer{}
}

func (s *dispatcherServer) SetTotalJobs(total int32) {
	s.mu.Lock()
	s.totalJobs = total
	s.mu.Unlock()
}

func (s *dispatcherServer) GetJob(ctx context.Context, req *primes.JobRequest) (*primes.Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case job, ok := <-jobs:
		if !ok {
			fmt.Println("No more jobs available.")
			return nil, fmt.Errorf("no more jobs available")
		}
		return &job, nil
	default:
		return nil, fmt.Errorf("no more jobs available")
	}
}

type consolidatorServer struct {
	primes.UnimplementedConsolidatorServer
	totalPrimes int32
	mu          sync.Mutex
}

func (s *consolidatorServer) Shutdown(ctx context.Context, req *primes.ShutdownRequest) (*primes.Ack, error) {
	log.Println("Shutting down the consolidator server.")
	os.Exit(0)
	return &primes.Ack{Success: true}, nil
}

func (s *dispatcherServer) setupClients(consolidatorAddress string) {
	var err error
	conn, err := grpc.Dial(consolidatorAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect to consolidator: %v", err)
	}
	s.consolidatorClient = primes.NewConsolidatorClient(conn)

}

func (s *consolidatorServer) SendResult(ctx context.Context, res *primes.Result) (*primes.Ack, error) {
	s.mu.Lock()
	s.totalPrimes += res.Count
	s.mu.Unlock()
	log.Printf("CONSOLIDATUS: Received result: %v, Total primes: %d", res, s.totalPrimes)
	return &primes.Ack{Success: true}, nil
}

func startServer(srv *grpc.Server, address string, serverName string) {
	// Setup network listener for the server
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", serverName, err)
	}

	reflection.Register(srv)

	log.Printf("%s is ready and listening on %s", serverName, address)

	// log if server fails
	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("failed to serve %s: %v", serverName, err)
		}
	}()
}

func (s *dispatcherServer) ReportCompletion(ctx context.Context, req *primes.JobCompletion) (*primes.Ack, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.completedJobs++

	log.Printf("DISPATCHIUM: Worker %s reported completion of it's job, prime count sent to consolidator. Total completed jobs: %d\n", req.WorkerId, s.completedJobs)

	allDone := s.completedJobs == s.totalJobs
	if allDone {
		elapsed := time.Since(s.startTime)
		fmt.Printf("\nAll jobs have been processed and all workers are inactive. Total time taken: %s\n", elapsed)
		fmt.Println("\n\nWorking on shutdown sequence, this will remain unimplemented because its not in project doc.")
		fmt.Println("Use Ctrl+C to close open servers cleanly, the go contexts are made to handle sigterm\n\n")
		// go s.sendShutdownSignal() // IN DEVELOPMENT
	}

	return &primes.Ack{
		Success: true,
		Message: fmt.Sprintf("Completion reported successfully for worker %s", req.WorkerId),
	}, nil
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "../primes_config.txt", "Path to configuration file")
	flag.Parse()

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	jobs = make(chan primes.Job, 100)

	dispatcher := &dispatcherServer{}
	dispatcherSrv := grpc.NewServer()
	primes.RegisterDispatcherServer(dispatcherSrv, dispatcher)
	startServer(dispatcherSrv, cfg.Dispatcher, "Dispatcher Server")

	dispatcher.setupConsolidatorClient(cfg.Consolidator)
	dispatcher.setupFileServerClient(cfg.FileServer)

	consolidator := &consolidatorServer{}
	consolidatorSrv := grpc.NewServer()
	primes.RegisterConsolidatorServer(consolidatorSrv, consolidator)
	startServer(consolidatorSrv, cfg.Consolidator, "Consolidator Server")

	dispatcher.startTime = time.Now() // start time when jobs are loaded
	totalJobs := loadJobs(*dataFilePath, *N)
	dispatcher.totalJobs = totalJobs

	select {}
}

func (s *dispatcherServer) setupConsolidatorClient(consolidatorAddress string) {
	conn, err := grpc.Dial(consolidatorAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect to consolidator: %v", err)
	}
	s.consolidatorClient = primes.NewConsolidatorClient(conn)
}

func (s *dispatcherServer) setupFileServerClient(fileserverAddress string) {
	conn, err := grpc.Dial(fileserverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect to fileserver: %v", err)
	}
	s.fileServerClient = primes.NewFileServerClient(conn)
}

func (s *dispatcherServer) sendShutdownSignal() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.consolidatorClient.Shutdown(ctx, &primes.ShutdownRequest{})
	if err != nil {
		log.Printf("Failed to send shutdown signal to consolidator: %v", err)
		return
	}
	log.Println("Shutdown signal sent successfully to consolidator.")
	_, err = s.fileServerClient.Shutdown(ctx, &primes.ShutdownRequest{})
	if err != nil {
		log.Printf("Failed to send shutdown signal to fileserver: %v", err)
		return
	}
	log.Println("Shutdown signal sent successfully to fileserver.")
}

func loadJobs(filePath string, jobSize int) int32 {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
		return 0
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		log.Fatalf("Failed to get file info: %v", err)
		return 0
	}

	var offset int64 = 0
	var totalJobs int32
	for offset < info.Size() {
		length := jobSize
		if offset+int64(length) > info.Size() {
			length = int(info.Size() - offset)
		}
		jobs <- primes.Job{
			Pathname: filePath,
			Start:    offset,
			Length:   int32(length),
		}
		offset += int64(length)
		totalJobs++
	}
	fmt.Println("Starting job distribution...")
	return totalJobs
}
