BIN := squares

.PHONY: all windows $(BIN) $(BIN).exe

all: $(BIN)

windows: $(BIN).exe

$(BIN):
	go build -o $@ -ldflags "-s -w"

$(BIN).exe:
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build -o $@ -tags static -ldflags "-s -w"
