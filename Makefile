.PHONY: build run test

build:
	@go build -o bin/fs

run: build
	@./bin/fs

test:
	@go test ./...
# 运行代码中所有单元测试，加上"-v"代表输出详情

