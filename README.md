# RunComfy serverless API Go bindings

[![Go Reference](https://pkg.go.dev/badge/github.com/iguanesolutions/go-runcomfyserverless.svg)](https://pkg.go.dev/github.com/iguanesolutions/go-runcomfyserverless) [![Go report card](https://goreportcard.com/badge/github.com/iguanesolutions/go-runcomfyserverless)](https://goreportcard.com/report/github.com/iguanesolutions/go-runcomfyserverless)

Go bindings to interract with [RunComfy](https://www.runcomfy.com/) [serverless API](https://www.runcomfy.com/comfyui-api).

Based on the official [API reference](https://docs.runcomfy.com/deployment-endpoints).

## Installation

```bash
go get -u github.com/iguanesolutions/go-runcomfyserverless
```

## Example

A complete code usage example can be found in the [example folder](https://github.com/iguanesolutions/go-runcomfyserverless/blob/main/example/main.go).
It executes a custom [Flux Kontext workflow](https://docs.comfy.org/tutorials/flux/flux-1-kontext-dev).

Execute it with:

```bash
go run example/main.go
```

### Generation

Note: lambda was warn

```text
Job sent: f033d089-3518-49c4-8abb-ef53018f7260
started ! (queued for 10.407347s)
still in progress
still in progress
still in progress
still in progress
still in progress
still in progress
Done (generation took 1m9.997048083s -not counting wait time-):
        ComfyUI_00021_.png
```

### Cancellation

```text
Job sent: b6f21678-6a84-4e0a-97f0-67563dbfca50
started ! (queued for 10.545094334s)
still in progress
still in progress
^C Cancelling request
Cancel request status: cancellation_requested
canceled
```
