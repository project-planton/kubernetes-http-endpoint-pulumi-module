.PHONY: deps
deps:
	go mod download
	go mod tidy

.PHONY: vet
vet:
	go vet ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: clean
clean:
	rm -rf ${build_dir}

.PHONY: build
build: clean deps vet fmt

.PHONY: update-deps
update-deps:
	go get buf.build/gen/go/plantoncloud/planton-cloud-apis/protocolbuffers/go@latest
	go get buf.build/gen/go/plantoncloud/planton-cloud-apis/grpc/go@latest
	go get github.com/plantoncloud-inc/pulumi-stack-runner-go-sdk
	go get github.com/plantoncloud-inc/kube-cluster-pulumi-blueprint
