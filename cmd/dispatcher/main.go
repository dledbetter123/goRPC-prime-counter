package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync/atomic"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/dledbetter123/distributed-prime-counter/pkg/config"
	primes "github.com/dledbetter123/distributed-prime-counter/pkg/primespb"
)

var (
	jobs          chan primes.Job
	dataFilePath  = flag.String("datafile", "../data/newgen.dat", "Path to the binary data file")
	activeWorkers int32
)

type dispatcherServer struct {
	primes.UnimplementedDispatcherServer
}

func (s *dispatcherServer) GetJob(ctx context.Context, req *primes.JobRequest) (*primes.Job, error) {
	select {
	case job, ok := <-jobs:
		if !ok {
			atomic.AddInt32(&activeWorkers, -1)
			if atomic.LoadInt32(&activeWorkers) == 0 {
				fmt.Println("All jobs have been processed and all workers are inactive.")
			}
			return nil, fmt.Errorf("no more jobs available")
		}
		return &job, nil
	default:
		return nil, fmt.Errorf("no jobs available")
	}
}

func startServer(srv *grpc.Server, address string) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	reflection.Register(srv)
	fmt.Println("Dispatcher server listening on", address)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func main() {
	flag.Parse()

	jobs = make(chan primes.Job, 100)

	cfg, err := config.LoadConfig("../primes_config.txt")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Println("Attempting to bind to address:", cfg.Dispatcher)

	go func() {
		loadJobs(*dataFilePath, 64*1024)
		close(jobs)
	}()

	dispatcherSrv := grpc.NewServer()
	primes.RegisterDispatcherServer(dispatcherSrv, &dispatcherServer{})
	startServer(dispatcherSrv, cfg.Dispatcher)
}

func loadJobs(filePath string, jobSize int) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		log.Fatalf("Failed to get file info: %v", err)
		return
	}

	var offset int64 = 0
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
	}
}
