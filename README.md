**:exclamation: This project is under construction and has not yet published a working version**

#  :bulb: feature backlog

## command line arguments

command line arguments should be parsed using an available library

--help [-h]         : print available command line arguments and explanation

--concurrency [-c]  : how many threads should be used to fire requests

--duration [-d]     : how long should the load test run

--once              : just execute a single scenario run, ignores concurrency and duration

--file [-f]         : the .http file to use, format uses a subset of intellij http syntax

--env [-e]          : file to use as variable input, format ``-e <file>:<env-key>``, ie.``-e client.env.json:service-dev``

--output [-o]       : output file/format, default is output format is human friendly tabular output, machine-readable options should be available with  "-o json" or "-o yaml"

## report format

- add reporter for json and yaml output
- maybe use interface for multiple reporters "func Report(r internal.Result)"
- -o syntax similar to kubectl

## executor suuport "once"

- the executor should allow running a scenario only once
- "once" can be used if the goal is not a load test, but just running the .http file (useful for general api-testing or automated smoke tests)

## excecutor support "duration"

- the executor should allow running multiple requests for a period of time
- use a minimal sleep between executions to not overwhelm the local process

## executor support "concurrency"

- the executor should allow execution on multiple threads inside go routines
- execution results should be returned from each go routine after finishing

## authentication support

- many mature systems require authentication to use their endpoints
- jetter should support a pre-execution hook to retrieve a jwt
- variables and credentials should be given via env file
- use the same or similar (but compatible) format like intellij env file
- authentication should be configured intellij compatible like this ``Authorization: Bearer {{$auth.token("auth-id")}}``

--> TESTING: need a local docker-compose and basic config for keycloak

## variable support from env file

- static variables defined in the env file should be replaced before execution

## "magic variable" support

- allow using magic variables like "{{$uuid}}"
- these variables will be automatically replaced before execution

| Variable | Output      |
|----------|-------------|
| {{$uuid}}  | random UUID |
| {{$tsid}} | random TSID |

## support for intellij request configuration

- intellij .http syntax allows using ͘``# @timeout 10`` and other configuration
- maybe not all directives can be supported or make sense, needs to be checked
- this should be added to the parser and the http client should pick up the configuration

## support global configuration directives

- allow global configuration which can be put at the top a .http file
- like ``#@jetter threshold_http_req_failed 0.01``