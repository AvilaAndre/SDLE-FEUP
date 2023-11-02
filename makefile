BIN = bin/
ENDPOINT = localhost

all: clean $(BIN)/database_node $(BIN)/orchestrator

debug_db_node:
	go run database_node/main.go

run_db_node: $(BIN)/database_node
	./$(BIN)/db_node $(CON_ENDPOINT) $(CON_PORT) $(DATA_PORT)

run_orchestrator: $(BIN)/orchestrator
	./$(BIN)/orchestrator $(CON_PORT)

$(BIN)/database_node:
	go build -o $(BIN)/db_node database_node/*

$(BIN)/orchestrator:
	go build -o $(BIN)/orchestrator orchestrator/*


.PHONY: clean
clean:
	rm -f $(BIN)/*



# make run_db_node CON_ENDPOINT=192.168.1.70 CON_PORT=6873 DATA_PORT=6878
# make run_orchestrator CON_PORT=6873