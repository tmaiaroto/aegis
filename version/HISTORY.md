This file will include a bit of history/release notes.
Note that it is not a complete accounting of changes. It's a weak attempt at a more formal process.

## version 0.x

Before verison 1.x, aegis focused primarily on being a deploy tool for AWS Lambda and API Gateway.
It used a Node.js shim to get Go to work with Lambda pre-native Go support. It had a router and
some helpers. That's about it. The goal was educational as well as just a quick way to start a Go
project/API using Lambda.

## verison 1.x

This version marks a major milestone. Native Go support was made available for Lambda and starting
with this new major version, the Node.js shim was ditched. The decision was also made to start
expanding upon the simple convenient functions and handlers/routers. This is the beginning of
a more proper "framework." The intent is a lightweight framework and deploy tool.

Aegis is opinionated and prefers convention over configuration. It provides many helpers to
handle incoming messages and invents to Lambda functions. It also provides some relevant
AWS infrastructure/resource creation and management.

The goal is not to create a feature rich infrastructure management tool. Use something like
Terraform. The goal is not to create a framework for use with multiple clouds that supports
as many services, languages, and providers as possible. Use something like Serverless.

The goal is to do a few things well instead of trying to do everything. It's about providing
a lightweight set of helpers or framework to help build things faster. It's to be conventional
and flexible. 1.x will focus on adding more event router/handlers and helper functions.
Not every possible service will likely ever covered, the focus will be on the common.

## 1.6.0

- S3 bucket notification triggers and router handler
- begin restructuring/organization of functions for deploy command (Deployer interface)