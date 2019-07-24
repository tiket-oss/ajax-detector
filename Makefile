TARGET = ajax-detector
SAMPLE_PAGE_URL = https://www.google.com

.PHONY: build
build:
	go build -o bin/$(TARGET) main/main.go

.PHONY: testrun
testrun: build
	@echo "Running $(TARGET) with default flags..."
	./bin/$(TARGET) $(SAMPLE_PAGE_URL)