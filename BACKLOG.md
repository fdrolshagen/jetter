#  :bulb: feature backlog

## command line arguments

--concurrency [-c]  : how many threads should be used to fire requests

--output [-o]       : output file/format, default is output format is human friendly tabular output, machine-readable options should be available with  "-o json" or "-o yaml"

## report format

- add reporter for json and yaml output
- maybe use interface for multiple reporters "func Report(r internal.Result)"
- -o syntax similar to kubectl

## executor support "concurrency"

- the executor should allow execution on multiple threads inside go routines
- execution results should be returned from each go routine after finishing

## token refresh after expiry
- if token expires during a scenario jetter should automatically refresh or obtain new token

## support for intellij request configuration

- intellij .http syntax allows using ͘``# @timeout 10`` and other configuration
- maybe not all directives can be supported or make sense, needs to be checked
- this should be added to the parser and the http client should pick up the configuration

## support jetter directives (per-request or global)

- allow global configuration which can be put at the top a .http file
- global ``#@jetter threshold_http_req_failed 0.01``
- per-request ``#@jetter extract ID $.username`` and in another request use ``{{$vars("ID")}}.``