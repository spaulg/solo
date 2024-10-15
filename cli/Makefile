GOCMD=go
GOTEST=$(GOCMD) test
GOCOVER=$(GOCMD) tool cover

.PHONY: test

test:
	$(GOTEST) -v ./...

cover:
	$(GOTEST) -coverprofile=coverage.out -v ./...
	$(GOCOVER) -html=coverage.out

coverage.out:
	make test
