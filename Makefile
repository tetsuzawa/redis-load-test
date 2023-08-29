GOARCH := arm64
GOOS := linux

.PHONY: deploy_all
deploy_all: bin/redis_inserter bin/redis_inserter_with_value bin/redis_inserter_set bin/reader bin/server
	rsync -avz $^ ssm-user@43.206.156.127:/home/ssm-user/

bin/redis_inserter: ./*.go cmd/redis_inserter
	GOARCH=$(GOARCH) GOOS=$(GOOS) go build -o $@ ./cmd/redis_inserter/
	chmod +x $@

bin/redis_inserter_with_value: ./*.go ./cmd/redis_inserter_with_value/*.go
	GOARCH=$(GOARCH) GOOS=$(GOOS) go build -o $@ ./cmd/redis_inserter_with_value/
	chmod +x $@

bin/redis_inserter_set: ./*.go ./cmd/redis_inserter_set/*.go
	GOARCH=$(GOARCH) GOOS=$(GOOS) go build -o $@ ./cmd/redis_inserter_set/
	chmod +x $@

bin/reader: ./*.go ./cmd/reader/*.go
	GOARCH=$(GOARCH) GOOS=$(GOOS) go build -o $@ ./cmd/reader/
	chmod +x $@


bin/server: ./*.go ./cmd/server/*.go
	GOARCH=$(GOARCH) GOOS=$(GOOS) go build -o $@ ./cmd/server/
	chmod +x $@

