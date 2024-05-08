cd ../cmd/fileserver
go build -o fileserver

cd ../cmd/dispatcher
go build -o dispatcher

cd ../cmd/consolidator
go build -o consolidator

cd ../cmd/worker
go build -o worker

