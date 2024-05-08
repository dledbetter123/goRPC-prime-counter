# Prime Finder in Binary Files

This Go program efficiently searches for prime numbers in binary files using a multithreaded approach. It divides the file into segments, assigns them to worker threads, and processes these segments concurrently to maximize resource utilization and minimize processing time.

## Prerequisites

- A binary file to process (see the section on generating test data)

## Usage

1. **Compiling the Program**

   First, compile the program using the Go toolchain:

   ```bash
   go build -o primefinder prime.go
   ```

   This command generates an executable named `primefinder`.

2. **Generating Test Data**

   The program includes a feature to generate a binary file filled with sequential numbers for testing purposes. To generate a file, run:

   ```bash
   ./primefinder -generate=true -min=<minimum value> -max=<maximum value> -rng=<number of values> -random=<random or sequential>
   ```

   By default, this generates a file named `newgen.dat`

3. **Running the Program**

   To run the program, use the following syntax:

   ```bash
   ./primefinder -pathname=<path to binary file> -M=<number of workers> -N=<segment size> -C=<chunk size>
   ```

   - `-pathname`: Path to the binary file to be processed.
   - `-M`: Number of worker threads to use for processing.
   - `-N`: Size of each segment in bytes. Each worker thread processes one segment at a time.
   - `-C`: Chunk size in bytes. Each read operation within a worker processes this amount of bytes.

   **Example:**

   ```bash
   ./primefinder -pathname=./data/newgen.dat -M=10 -N=65536 -C=1024
   ```

## Configuration Flags

### Number Generation

- `-generate`: Set to `true` to activate data file generation mode.
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

(base) davidledbetter@Davids-MacBook-Pro-4 golang-multiThread-prime % go run prime.go -pathname 20_million_s.dat -generate -min=0 -max=20000000 -rng=20000000 -random=false 
20_million_s.dat Data file generated, min=0, max=20000000, rng=20000000 Note the values generated are WITH replacement when random is enabled.
This was tested on ranges 1,000-1 billion to baseline accurate prime counting
```

### Prime Calculation

- `-pathname`: Required. Specifies the path to the binary file to process.
- `-M`: Number of worker threads (`1` by default).
- `-N`: Size of each segment in bytes (`65536` by default).
- `-C`: Chunk size in bytes for each read operation (`1024` by default).

## Output

The program logs the progress and results of the processing to the console. This includes the number of prime numbers found in each segment and the total number of primes found in the entire file. this is running on a (non-random) range from 0-500,000

```
(base) davidledbetter@Davids-MacBook-Pro-4 golang-multiThread-prime % go run prime.go -pathname newgen.dat -M 50
2024/03/03 21:50:41 Worker 30 completed job: {Pathname:newgen.dat Start:0 Length:65536} with result: {Job:{Pathname:newgen.dat Start:0 Length:65536} Count:1028}
2024/03/03 21:50:41 Worker 45 completed job: {Pathname:newgen.dat Start:65536 Length:65536} with result: {Job:{Pathname:newgen.dat Start:65536 Length:65536} Count:872}
2024/03/03 21:50:41 Worker 30 completed job: {Pathname:newgen.dat Start:131072 Length:65536} with result: {Job:{Pathname:newgen.dat Start:131072 Length:65536} Count:825}
...
2024/03/03 21:50:41 Worker 45 completed job: {Pathname:newgen.dat Start:3735552 Length:65536} with result: {Job:{Pathname:newgen.dat Start:3735552 Length:65536} Count:647}
2024/03/03 21:50:41 Worker 20 completed job: {Pathname:newgen.dat Start:3866624 Length:65536} with result: {Job:{Pathname:newgen.dat Start:3866624 Length:65536} Count:619}
2024/03/03 21:50:41 Worker 11 completed job: {Pathname:newgen.dat Start:3801088 Length:65536} with result: {Job:{Pathname:newgen.dat Start:3801088 Length:65536} Count:637}
2024/03/03 21:50:41 Worker 43 completed job: {Pathname:newgen.dat Start:3932160 Length:65536} with result: {Job:{Pathname:newgen.dat Start:3932160 Length:65536} Count:636}
Total primes: 41538

```

# to register the package name instead of using go run prime.go or building.

```

(base) davidledbetter@Davids-MacBook-Pro-4 golang-multiThread-prime % go install prime/david/prime
(base) davidledbetter@Davids-MacBook-Pro-4 golang-multiThread-prime % go install .                 
(base) davidledbetter@Davids-MacBook-Pro-4 golang-multiThread-prime % go install  
(base) davidledbetter@Davids-MacBook-Pro-4 golang-multiThread-prime % export PATH=$PATH:$(dirname $(go list -f '{{.Target}}' .))

(base) davidledbetter@Davids-MacBook-Pro-4 golang-multiThread-prime % prime -pathname million_s.dat -M 50
0x14000104260
million_s.dat
2024/03/04 00:28:26 Worker 2 completed job: {Pathname:million_s.dat Start:0 Length:65536} with result: {Job:{Pathname:million_s.dat Start:0 Length:65536} Count:1028}
2024/03/04 00:28:26 Worker 2 completed job: {Pathname:million_s.dat Start:65536 Length:65536} with result: {Job:{Pathname:million_s.dat Start:65536 Length:65536} Count:872}
2024/03/04 00:28:26 Worker 2 completed job: {Pathname:million_s.dat Start:131072 Length:65536} with result: {Job:{Pathname:million_s.dat Start:131072 Length:65536} Count:825}
2024/03/04 00:28:26 Worker 2 completed job: {Pathname:million_s.dat Start:196608 Length:65536} with result: {Job:{Pathname:million_s.dat Start:196608 Length:65536} Count:787}
2024/03/04 00:28:26 Worker 40 completed job: {Pathname:million_s.dat Start:262144 Length:65536} with result: {Job:{Pathname:million_s.dat Start:262144 Length:65536} Count:776}
...
2024/03/04 00:28:26 Worker 40 completed job: {Pathname:million_s.dat Start:7798784 Length:65536} with result: {Job:{Pathname:million_s.dat Start:7798784 Length:65536} Count:571}
2024/03/04 00:28:26 Worker 42 completed job: {Pathname:million_s.dat Start:7864320 Length:65536} with result: {Job:{Pathname:million_s.dat Start:7864320 Length:65536} Count:590}
2024/03/04 00:28:26 Worker 10 completed job: {Pathname:million_s.dat Start:7929856 Length:65536} with result: {Job:{Pathname:million_s.dat Start:7929856 Length:65536} Count:591}
Total primes: 78498
Elasped Time (milliseconds): 588

```