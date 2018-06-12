# Init

The first CLI command you may use, especially when getting started, is `aegis init`. This will
create a `main.go` and `aegis.yaml` file in the current directory.

_Note that if you already have either of these files present, the command will not replace or
touch your existing files._

The contents of `main.go` include a short boilerplate example Aegis application with API Gateway
event handling via the `Router`.

The `aegis.yaml` then includes configuration for creating an API Gateway upon deploy.

This is your basic "hello world" type example and it's designed to get you started quickly.
You need not use this if you don't want, but often times it's saves a few steps. You'll likely
run the command then go into the `aegis.yaml` file to change some names around.
