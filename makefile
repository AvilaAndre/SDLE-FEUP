BIN = bin
ENDPOINT = localhost

all: clean $(BIN)/database_node $(BIN)/load_balancer 

run_db_node: $(BIN)/database_node
	./$(BIN)/database_node $(OWN_PORT) $(BAL_ADDR) $(BAL_PORT)

run_load_balancer: $(BIN)/load_balancer
	./$(BIN)/load_balancer $(OWN_PORT)

$(BIN)/database_node:
	go build -o $@ ./database_node

$(BIN)/load_balancer:
	go build -o $@ ./load_balancer

.PHONY: clean
clean:
	rm -f $(BIN)/*
