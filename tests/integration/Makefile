test:
	go test -v ./...

lint:
	golangci-lint run

bench:
	go test -bench=BenchmarkRead -benchtime=100x -cpu=1 -run=^$$ ./performance