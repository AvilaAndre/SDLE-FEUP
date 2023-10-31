BIN = bin/
ENDPOINT = localhost

all: clean $(BIN)/database_node $(BIN)/orchestrator

debug_db_node:
	go run database_node/main.go

run_db_node: $(BIN)/database_node
	./$(BIN)/db_node $(ENDPOINT) $(PORT)

run_orchestrator: $(BIN)/orchestrator
	./$(BIN)/orchestrator $(PORT)

$(BIN)/database_node:
	go build -o $(BIN)/db_node database_node/*

$(BIN)/orchestrator:
	go build -o $(BIN)/orchestrator orchestrator/*


.PHONY: clean
clean:
	rm -f $(BIN)/*