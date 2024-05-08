package main

import (
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dledbetter123/distributed-prime-counter/pkg/config"
	primes "github.com/dledbetter123/distributed-prime-counter/pkg/primespb"
	"google.golang.org/grpc"
)

var (
	configPath = flag.String("config", "../primes_config.txt", "Path to the configuration file")
	chunkSize  = flag.Int("C", 1024, "Chunk size in bytes")
	workerID   = flag.String("id", "worker-"+time.Now().Format("20060102150405"), "Worker ID")
)

func connectToDispatcher(address string) (*grpc.ClientConn, primes.DispatcherClient) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect to Dispatcher: %v", err)
	}
	client := primes.NewDispatcherClient(conn)
	return conn, client
}

func connectToFileServer(address string) (*grpc.ClientConn, primes.FileServerClient) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect to FileServer: %v", err)
	}
	client := primes.NewFileServerClient(conn)
	return conn, client
}

func connectToConsolidator(address string) (*grpc.ClientConn, primes.ConsolidatorClient) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect to Consolidator: %v", err)
	}
	client := primes.NewConsolidatorClient(conn)
	return conn, client
}

func handleSignals(cancel context.CancelFunc) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	cancel() // Cancel context on receiving a signal
}

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go handleSignals(cancel) // Go routine to listen for interrupt signals

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	connDisp, dispatcherClient := connectToDispatcher(cfg.Dispatcher)
	defer connDisp.Close()

	connFile, fileClient := connectToFileServer(cfg.FileServer)
	defer connFile.Close()

	connCons, consolidatorClient := connectToConsolidator(cfg.Consolidator)
	defer connCons.Close()

	worker(ctx, dispatcherClient, fileClient, consolidatorClient, cancel)
	<-ctx.Done()
	log.Println("Worker shutting down gracefully.")
}

func worker(ctx context.Context, dispatcher primes.DispatcherClient, fileserver primes.FileServerClient, consolidator primes.ConsolidatorClient, cancel context.CancelFunc) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, worker is shutting down.")
			return
		default:
			job, err := dispatcher.GetJob(ctx, &primes.JobRequest{WorkerId: *workerID})
			if err != nil {
				if strings.Contains(err.Error(), "no more jobs available") {
					log.Printf("Worker %s: No more jobs available, shutting down.", *workerID)
					cancel()
					return // Correctly exit the loop and shut down the worker
				}
				log.Printf("Worker %s: Failed to get job: %v", *workerID, err)
				time.Sleep(time.Second) // Prevent spamming requests in case of persistent error
				continue
			}

			// Process the job as usual
			if err := processAndSendResults(ctx, fileserver, consolidator, job); err != nil {
				log.Printf("Worker %s: Error while processing job: %v", *workerID, err)
			}
		}
	}
}

func processAndSendResults(ctx context.Context, fileserver primes.FileServerClient, consolidator primes.ConsolidatorClient, job *primes.Job) error {
	stream, err := fileserver.GetFileChunk(ctx, &primes.JobRequest{Job: job})
	if err != nil {
		return fmt.Errorf("failed to start file chunk stream: %v", err)
	}

	primesCount := 0
	buffer := make([]byte, *chunkSize)
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to receive chunk: %v", err)
		}

		copy(buffer, chunk.ChunkData)
		primesCount += countPrimes(buffer)
	}

	_, err = consolidator.SendResult(ctx, &primes.Result{Count: int32(primesCount)})
	if err != nil {
		return fmt.Errorf("failed to send result: %v", err)
	}

	log.Printf("Worker %s completed job with result: primes=%d", *workerID, primesCount)
	return nil
}

func isPrime(n uint64) bool {
	if n <= 1 {
		return false
	}
	if n <= 3 {
		return true
	}
	if n%2 == 0 || n%3 == 0 {
		return false
	}
	for i := uint64(5); i*i <= n; i += 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return false
		}
	}
	return true
}

func countPrimes(data []byte) int {
	primeCount := 0
	for i := 0; i+8 <= len(data); i += 8 {
		num := binary.LittleEndian.Uint64(data[i : i+8])
		if isPrime(num) {
			primeCount++
		}
	}
	return primeCount
}
