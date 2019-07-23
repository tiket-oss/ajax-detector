TARGET = page-profile
SAMPLE_PAGE_URL = https://www.tiket.com

.PHONY: build
build:
	go build -o bin/$(TARGET)

.PHONY: testrun
testrun: build
	@echo "Running $(TARGET) with default flags..."
	./bin/$(TARGET) $(SAMPLE_PAGE_URL)