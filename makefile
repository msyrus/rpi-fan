GOCMD=go
BINARY_NAME=rpi-fan
GOOS=linux
GOARCH=arm

BUILD_DIR=build
BINARY=$(BUILD_DIR)/$(BINARY_NAME)

.PHONY: all clean run install uninstall

all: $(BINARY)

clean:
	$(GOCMD) clean
	rm -f $(BINARY)

$(BINARY): main.go
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOCMD) build -o $(BINARY) -v main.go

run: $(BINARY)
	$(BINARY)

install: $(BINARY) $(BINARY_NAME).service
	cp $(BINARY) /usr/bin/
	cp $(BINARY_NAME).service /etc/systemd/system/

uninstall:
	rm -f /usr/bin/$(BINARY_NAME)
	rm -f /etc/systemd/system/$(BINARY_NAME).service
