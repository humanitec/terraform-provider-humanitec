# Terraform Provider Humanitec

* [Usage](https://registry.terraform.io/providers/humanitec/humanitec/latest)
* [Documentation](https://registry.terraform.io/providers/humanitec/humanitec/latest/docs)

## Requirements

* [Terraform](https://www.terraform.io/downloads.html) >= 1.0
* [Go](https://golang.org/doc/install) >= 1.23

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the provider

Fill this in for each provider

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```

Once changes are merged, a new release can be created through the [new releases](https://github.com/humanitec/terraform-provider-humanitec/releases/new) page. We use `v` in front of a semantic version and generate release notes using the button in Github.
