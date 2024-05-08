#!/bin/bash

CONFIG="/Users/davidledbetter/CMSC621/distributed-prime-counter/config/primes_config.txt"
DATA_FILE_PATH="/Users/davidledbetter/CMSC621/distributed-prime-counter/newgen.dat"
PIPELINE_PID=0
FILESERVER_PID=0

trap "echo 'Shutting down...'; kill $PIPELINE_PID $FILESERVER_PID; exit" INT TERM TSTP

echo "Starting Dispatcher-Consolidator..."
cd cmd/pipeline
go run main.go -config "$CONFIG" -datafile "$DATA_FILE_PATH" -N 131072 &
PIPELINE_PID=$!
cd -

echo "Starting Fileserver..."
cd cmd/fileserver
go run main.go -config "$CONFIG" -datafile "$DATA_FILE_PATH" -C 2048 &
FILESERVER_PID=$!
cd -

sleep 2

echo "Starting Workers..."
cd cmd/worker

# set this for number of workers
for i in {1..8}; do
    WORKER_ID="worker-$i"
    go run main.go -config $CONFIG -id $WORKER_ID -C 2048 &
    echo "Starting worker with ID $WORKER_ID"
done

sleep 2
cd -

echo "jobs done, and workers put to rest, press ^C to exit."

# Wait for all processes to finish
wait