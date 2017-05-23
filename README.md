# go-aws-parallel
Scripts to do a lot of AWS API calls in parallel. First use case: repairing S3 object ACLs.

Usage only for developers so far, because `main.go` is used as the configuration file where you specify what to do.

I think that is actually more comfortable than passing many parameters to a command-line tool.
