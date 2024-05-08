#!/bin/bash

CONFIG="../../primes_config.txt"

# Initialize PID variables with placeholder values
DISPATCHER_PID=0
FILESERVER_PID=0
CONSOLIDATOR_PID=0

# Trap both SIGINT (Ctrl+C) and SIGTSTP (Ctrl+Z) to handle user interruption
trap "echo 'Shutting down...'; kill $DISPATCHER_PID $FILESERVER_PID $CONSOLIDATOR_PID; exit" INT TERM TSTP

echo "Starting Dispatcher..."
cd cmd/dispatcher
go run main.go &
DISPATCHER_PID=$!
cd -

echo "Starting Fileserver..."
cd cmd/fileserver
go run main.go &
FILESERVER_PID=$!
cd -

echo "Starting Consolidator..."
cd cmd/consolidator
go run main.go &
CONSOLIDATOR_PID=$!
cd -

sleep 2

echo "Starting Workers..."
cd cmd/worker
for i in {1..5}; do
    WORKER_ID="worker-$i"
    go run main.go -config $CONFIG -id $WORKER_ID &
    echo "Starting worker with ID $WORKER_ID"
done

sleep 2
cd -

# Wait for all processes to finish
wait
