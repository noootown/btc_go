BINARY := blockchain

build:
	@echo "====> Go build"
	@go build -o $(BINARY)
