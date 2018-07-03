terraform-provider-vsts: main.go provider.go resource_project.go apiclient.go  resource_repository.go
	go build -o $@

clean:
	rm -f terraform-provider-vsts

test: clean terraform-provider-vsts
	terraform plan
