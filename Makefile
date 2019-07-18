TARGET = page-profiler
SAMPLE_PAGE_URL = https://www.tiket.com

.PHONY: build
build:
	go build -o bin/$(TARGET)

.PHONY: start
start:
	@echo "Starting $(TARGET) with default flags"
	./bin/$(TARGET) -page-url "$(SAMPLE_PAGE_URL)"