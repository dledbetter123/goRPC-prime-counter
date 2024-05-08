# Prime Finder in Binary Files

This Go program efficiently searches for prime numbers in binary files using gRPC approach. It divides the file into segments, assigns them to independent worker processes, and processes these segments concurrently to maximize resource utilization and minimize processing time.

![Alt text](screenshot.png?raw=true "Diagram")

## Prerequisites

- A binary file to process (see the section on generating test data)

## Configuration Flags

### Number Generation

- `-generate`: Set to activate data file generation mode.
- `-min`:the minimum value (inclusive) that can be written to the file. Default is `0`.
- `-max`: the maximum value (inclusive) that can be written to the file. Default is `1,000,000`.
- `-rng`: how many values to write to the file.
- `-random`: whether the values should be generated randomly (`true`) or sequentially (`false`). Default is `true`.

### Usage

To generate a binary file with random numbers within a specified range, use the following command format:

```bash
go run prime.go -generate -min=100 -max=1000000 -rng=500000 -random=false
Output:
Data file generated, min=100, max=1000000, samples=500000 Note the values generated are WITH replacement if random is enabled.
This was tested on ranges 1,000-1 billion to baseline accurate prime counting

(base) davidledbetter@Davids-MacBook-Pro-4 golang-multiThread-prime % go run prime.go -pathname 20_million_s.dat -generate -min=0 -max=20000000 -rng=20000000 -random=true 
20_million_s.dat Data file generated, min=0, max=20000000, rng=20000000 Note the values generated are WITH replacement when random is enabled.
This was tested on ranges 1,000-1 billion to baseline accurate prime counting
```

### Prime Calculation with startup script to use gRPC solution

- `-N`: Size of each segment in bytes `65536` by default. was also run with 128 * 1024 bytes, and 256 * 1024 bytes
- `-C`: Chunk size in bytes for each read operation (`1024` by default) was also run with 2048 bytes, and 8192 bytes.
```bash


#!/bin/bash

CONFIG="/Users/davidledbetter/CMSC621/distributed-prime-counter/config/primes_config.txt"
DATA_FILE_PATH="/Users/davidledbetter/CMSC621/distributed-prime-counter/prime_data/1_million_s.dat"
PIPELINE_PID=0
FILESERVER_PID=0

trap "echo 'Shutting down...'; kill $PIPELINE_PID $FILESERVER_PID; exit" INT TERM TSTP

echo "Starting Dispatcher-Consolidator..."
cd cmd/pipeline
go run main.go -config "$CONFIG" -datafile "$DATA_FILE_PATH" -N 65536 &
PIPELINE_PID=$!
cd -

echo "Starting Fileserver..."
cd cmd/fileserver
go run main.go -config "$CONFIG" -datafile "$DATA_FILE_PATH" -C 1024 &
FILESERVER_PID=$!
cd -

sleep 2

echo "Starting Workers..."
cd cmd/worker

# set this for number of workers
for i in {1..4}; do
    WORKER_ID="worker-$i"
    go run main.go -config $CONFIG -id $WORKER_ID -C 1024 &
    echo "Starting worker with ID $WORKER_ID"
done

sleep 2
cd -

echo "jobs done, and workers put to rest, press ^C to exit."

# Wait for all processes to finish
wait
```

## Output

The program logs the progress and results of the processing to the console. This includes the number of prime numbers found in each job and the total number of primes found in the entire file. this is running on a (non-random) range from 0-1 million

```
(base) davidledbetter@Davids-MacBook-Pro-4 golang-multiThread-prime % go run prime.go -pathname newgen.dat -M 50
(base) davidledbetter@Davids-MBP-4 distributed-prime-counter % ./start_system.sh
Starting Dispatcher-Consolidator...
/Users/davidledbetter/CMSC621/distributed-prime-counter
Starting Fileserver...
/Users/davidledbetter/CMSC621/distributed-prime-counter
Starting fileserver, managing datafile: /Users/davidledbetter/CMSC621/distributed-prime-counter/prime_data/1_million_s.dat
Starting fileserver, listening on localhost:5003
2024/05/08 14:01:01 Dispatcher Server is ready and listening on localhost:5001
2024/05/08 14:01:01 Consolidator Server is ready and listening on localhost:5002
Starting Workers...
Starting worker with ID worker-1
Starting worker with ID worker-2
Starting worker with ID worker-3
Starting worker with ID worker-4
2024/05/08 14:01:03 CONSOLIDATUS: Received result: jobId:"/Users/davidledbetter/CMSC621/distributed-prime-counter/prime_data/1_million_s.dat"  count:1028, Total primes: 1028
2024/05/08 14:01:03 Worker worker-3 completed job with result: primes=1028
2024/05/08 14:01:03 DISPATCHIUM: Worker worker-3 reported completion of it's job, prime count sent to consolidator. Total completed jobs: 1
2024/05/08 14:01:03 CONSOLIDATUS: Received result: jobId:"/Users/davidledbetter/CMSC621/distributed-prime-counter/prime_data/1_million_s.dat"  count:872, Total primes: 1900
2024/05/08 14:01:03 Worker worker-3 completed job with result: primes=872
.
.
.
2024/05/08 14:01:03 CONSOLIDATUS: Received result: jobId:"/Users/davidledbetter/CMSC621/distributed-prime-counter/prime_data/1_million_s.dat"  count:591, Total primes: 78460
2024/05/08 14:01:03 Worker worker-3 completed job with result: primes=591
2024/05/08 14:01:03 DISPATCHIUM: Worker worker-3 reported completion of it's job, prime count sent to consolidator. Total completed jobs: 122
2024/05/08 14:01:03 CONSOLIDATUS: Received result: jobId:"/Users/davidledbetter/CMSC621/distributed-prime-counter/prime_data/1_million_s.dat"  count:38, Total primes: 78498
2024/05/08 14:01:03 Worker worker-3 completed job with result: primes=38
2024/05/08 14:01:03 DISPATCHIUM: Worker worker-3 reported completion of it's job, prime count sent to consolidator. Total completed jobs: 123

All jobs have been processed and all workers are inactive. Total time taken: 1.99907725s


Working on shutdown sequence, this will remain unimplemented because its not in project doc.
Use Ctrl+C to close open servers cleanly, the go contexts are made to handle sigterm


2024/05/08 14:01:03 Worker worker-3: No more jobs available, shutting down.
2024/05/08 14:01:03 Worker shutting down gracefully.
2024/05/08 14:01:03 Worker worker-4: No more jobs available, shutting down.
2024/05/08 14:01:03 Worker shutting down gracefully.
2024/05/08 14:01:03 Worker worker-2: No more jobs available, shutting down.
2024/05/08 14:01:03 Worker shutting down gracefully.
2024/05/08 14:01:04 Worker worker-1: No more jobs available, shutting down.
2024/05/08 14:01:04 Worker shutting down gracefully.
/Users/davidledbetter/CMSC621/distributed-prime-counter
jobs done, and workers put to rest, press ^C to exit.
^CShutting down...
Shutting down...
(base) davidledbetter@Davids-MBP-4 distributed-prime-counter %

```
