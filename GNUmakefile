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
	go build -o ${GOBIN}/terraform-provider-aztools #-o ~/.terraform.d/plugins/github.com/glueckkanja/aztools/0.7.0/linux_amd64/terraform-provider-aztools

init:
	cd ./examples && terraform init -upgrade

test: generate build init
	cd ./examples && terraform plan

plan: generate build
	cd ./examples && rm -f .terraform.lock.hcl && terraform init -upgrade && terraform plan

apply:
	cd ./examples && terraform apply -auto-approve

destroy:
	cd ./examples && terraform destroy -auto-approve
