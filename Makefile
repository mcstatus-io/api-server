# Variables
BINARY := bin/main
SOURCES := $(wildcard src/*.go)

# Build for the current platform
build:
	go build -o $(BINARY) $(SOURCES)

# Build for Linux
build-linux: GOOS := linux
build-linux: EXTENSION := 
build-linux: build-cross

# Build for Windows
build-windows: GOOS := windows
build-windows: EXTENSION := .exe
build-windows: build-cross

# Cross-compile for a specific platform (used by build-linux and build-windows)
build-cross: export GOOS := $(GOOS)
build-cross: BINARY := $(BINARY)$(EXTENSION)
build-cross: $(BINARY)

# Run the application
run: build
	./$(BINARY)

# Clean up generated files
.PHONY: clean
clean:
	rm -f $(BINARY) $(BINARY).exe