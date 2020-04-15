[![Actions Status](https://github.com/leebradley/myretail/workflows/Go/badge.svg)](https://github.com/leebradley/myretail/actions)

# Description

This is an example retail service for fetching and storing product data.

# Development

## Prerequisites
* Terraform - https://learn.hashicorp.com/terraform/getting-started/install.html
* Golang - https://golang.org/doc/install
* Google Cloud SDK - https://cloud.google.com/sdk/install

## Setup

### Google Cloud

After installing the prerequisites, setup a Google Cloud project. Make sure you have the project and the region setup correctly. For example:

```
gcloud config set project target-myretail-demo
```

```
gcloud config set run/region us-central1
```

Activate Google Cloud functions in the dashboard, as well as Google Cloud Storage. Make sure billing is enabled.

Activate the datastore:
https://console.cloud.google.com/datastore/setup?project=PROJECT-NAME

### Golang

Make sure your vendor dependencies are loaded locally:

```
cd src
go mod vendor
```

### Terraform

Create a tfvars file conforming with the variable definition in variables.tf. For example:
```
project_id="someretail-demo"
datastore_id="products"
region="us-central1"
bucket_name="someretail-demo-bucket"
```

Initialize terraform:

```
terraform init
```

To deploy:

```
terraform apply --var-file=YOURFILE.tfvars
```

## Common commands

Get test coverage:

```
cd src
go test -coverprofile cp.out
go tool cover -html=cp.out
```

See errors:

```
gcloud functions logs read products
```

## Missing features / TODO

* Currently the PUT endpoint is exposed. Ideally endpoints can be managed with IAM roles. Google Cloud Function endpoint functionality is experimental, but it should be possible.
* Abstract logging so that can be unit tested (this will also clean up unit test output)

## Credits

This is based on / inspired by:
* https://github.com/terraform-google-modules/terraform-google-event-function
* https://github.com/googleapis/google-cloud-go/blob/master/datastore/datastore.go
* https://medium.com/swlh/using-pure-golang-for-google-cloud-bacc6b62e0ed
* https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body
* https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779