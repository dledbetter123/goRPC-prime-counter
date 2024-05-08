package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

var (
	pathname = flag.String("pathname", "newgen.dat", "Path to the binary file")
	M        = flag.Int("M", 1, "Number of worker threads")
	N        = flag.Int("N", 64*1024, "Size of each segment in bytes")
	C        = flag.Int("C", 1024, "Chunk size in bytes")
)

type Job struct {
	Pathname string
	Start    int64
	Length   int
}

type Result struct {
	Job   Job
	Count int
}

func isPrime(n uint64) bool {
	if n <= 1 {
		return false
	}
	for i := uint64(2); i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func worker(id int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	randSleep := time.Millisecond * time.Duration(rand.Intn(200)+400)
	time.Sleep(randSleep)
	for job := range jobs {
		file, err := os.Open(job.Pathname)
		if err != nil {
			log.Fatalf("Worker %d: Failed to open file: %v", id, err)
		}
		defer file.Close()

		primesCount := 0
		buffer := make([]byte, *C)
		for i := 0; i < job.Length; i += *C {
			readSize := *C
			if job.Length-i < *C {
				readSize = job.Length - i
			}
			_, err := file.ReadAt(buffer[:readSize], job.Start+int64(i))
			if err != nil {
				log.Fatalf("Worker %d: Failed to read file: %v", id, err)
			}

			for j := 0; j < readSize; j += 8 {
				// messed up using this because it doesn't read a 'uint64' value from the buffer
				// binary.LittleEndian.PutUint64(buffer[j:j+8], num)
				// this does
				num := binary.LittleEndian.Uint64(buffer[j : j+8])
				if isPrime(num) {
					primesCount++
				}
			}
		}

		result := Result{Job: job, Count: primesCount}
		results <- result
		log.Printf("Worker %d completed job: %+v with result: %+v\n", id, job, result)
	}
}

func dispatcher(pathname string, fileSize int64, jobs chan<- Job) {
	for start := int64(0); start < fileSize; start += int64(*N) {
		length := *N
		if start+int64(length) > fileSize {
			length = int(fileSize - start)
		}
		job := Job{Pathname: pathname, Start: start, Length: length}
		jobs <- job
	}
	close(jobs)
}

func consolidator(results <-chan Result, done chan<- int) {
	totalPrimes := 0
	for result := range results {
		totalPrimes += result.Count
	}
	done <- totalPrimes
}

var (
	generateData = flag.Bool("generate", false, " binary data file for testing")
	mini         = flag.Uint64("min", 0, "Minimum value to write, default 0")
	max          = flag.Uint64("max", 1000000, "maximum value to write, default 1,000,000")
	rng          = flag.Uint64("rng", 1000000, "how many values to write")
	randomize    = flag.Bool("random", true, "Generate values randomly [T] or sequentially ")
)

func generateBinaryFile(filename string, mini uint64, max uint64, numSamples uint64, randomize bool) {
	if mini > max {
		log.Fatalf("Minimum value cannot be greater than maximum value")
	}

	maxSamples := max - mini + 1
	if numSamples > maxSamples {
		log.Fatalf("Number of samples (%d) cannot exceed the range", numSamples)
	}

	binaryFile, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Cannot create binary file: %v", err)
	}
	defer binaryFile.Close()

	readmeFile, err := os.Create("readmeGEN.txt")
	if err != nil {
		log.Fatalf("Cannot create text file: %v", err)
	}
	defer readmeFile.Close()

	// rand.Seed(time.Now().UnixNano())

	for i := uint64(0); i < numSamples; i++ {
		var num uint64
		if randomize {
			num = mini + uint64(rand.Int63n(int64(max-mini+1)))
		} else {
			// make sure we don't exceed max
			num = mini + (i % (max - mini + 1))
		}

		if err := binary.Write(binaryFile, binary.LittleEndian, num); err != nil {
			log.Fatalf("Failed to write to binary file: %v", err)
		}

		if i < 100 {
			_, err := readmeFile.WriteString(fmt.Sprintf("%d\n", num))
			if err != nil {
				log.Fatalf("Failed to write to readme file: %v", err)
			}
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	flag.Parse()

	/*               GENERATION!!!!!             */
	if *generateData {

		generateBinaryFile(*pathname, *mini, *max, *rng, *randomize) // file going up to 500,000
		fmt.Printf("%s Data file generated, min=%d, max=%d, rng=%d Note the values generated are WITH replacement when random is enabled.\n", *pathname, *mini, *max, *rng)
		fmt.Println("This was tested on ranges 1,000-1 billion to baseline accurate prime counting")
		return
	}
	/*               GENERATION!!!!!             */

	/*					  v						 */

	/*               CALCULATION!!!!!             */
	fmt.Println(pathname)
	fmt.Println(*pathname)
	if *pathname == "" {
		log.Fatal("Pathname is required")
	}

	fileInfo, err := os.Stat(*pathname)
	if err != nil {
		log.Fatalf("Failed to get file stats: %v", err)
	}

	timer := time.Now().UnixMilli()

	jobs := make(chan Job, 100)       // job queue can hold up to 100 jobs
	results := make(chan Result, 100) // queue entering consolidator can hold up to 100 results.
	done := make(chan int)

	var wg sync.WaitGroup
	for i := 0; i < *M; i++ {
		wg.Add(1)
		go worker(i, jobs, results, &wg)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	go dispatcher(*pathname, fileInfo.Size(), jobs)
	go consolidator(results, done)

	totalPrimes := <-done
	timer = time.Now().UnixMilli() - timer
	fmt.Printf("Total primes: %d\n", totalPrimes)
	fmt.Printf("Elasped Time (milliseconds): %d\n", timer)
	/*               CALCULATION!!!!!             */
}
