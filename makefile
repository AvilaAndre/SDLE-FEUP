BIN = bin/

all: build

debug_db_node:
	go run database_node/main.go

run_db_node:
	./$(BIN)/db_node

run_orchestrator:
	./$(BIN)/orchestrator

.PHONY: build_db_node
build_db_node:
	go build -o $(BIN)/db_node database_node/main.go

.PHONY: build_orchestrator
build_orchestrator:
	go build -o $(BIN)/orchestrator orchestrator/main.go


.PHONY: build
build: build_db_node build_orchestrator

.PHONY: clean
clean:
	rm -f $(BIN)/*