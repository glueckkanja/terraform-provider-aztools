#default: testacc
#
## Run acceptance tests
#.PHONY: testacc
#testacc:
#	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

# Dev
default: test

generate:
	go generate
	go fmt

build:
	go build -o ${GOBIN}/terraform-provider-aztools

test: generate build
	cd ./examples && rm -f .terraform.lock.hcl && terraform init -upgrade && terraform apply -auto-approve

plan: generate build
	cd ./examples && rm -f .terraform.lock.hcl && terraform init -upgrade && terraform plan

apply:
	cd ./examples && terraform apply -auto-approve

destroy:
	cd ./examples && terraform destroy -auto-approve
