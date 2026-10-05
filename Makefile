gentest:
	rm -f ./testing/dag_json_*gen*.go
	go run ./testgen/main.go
.PHONY: gentest

test: gentest
	go test ./...
.PHONY: test

bench: gentest
	go test -bench=. -run='^$$' ./...
.PHONY: bench
