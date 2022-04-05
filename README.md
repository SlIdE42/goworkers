# goworkers

[![GitHub Workflow Status](https://github.com/SlIdE42/goworkers/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/SlIdE42/goworkers/actions/workflows/ci.yml?query=branch:master)
[![codecov](https://codecov.io/gh/SlIdE42/goworkers/branch/master/graph/badge.svg)](https://codecov.io/gh/SlIdE42/goworkers)
[![golangci](https://golangci.com/badges/github.com/SlIdE42/goworkers.svg)](https://golangci.com/r/github.com/SlIdE42/goworkers)
[![GoDoc](https://img.shields.io/badge/pkg.go.dev-doc-blue)](http://pkg.go.dev/github.com/SlIdE42/goworkers)
[![Go Report Card](https://goreportcard.com/badge/github.com/SlIdE42/goworkers)](https://goreportcard.com/report/github.com/SlIdE42/goworkers)

goworkers is a simple but effecient dynamic workers pool

## Types

### type [Pool](/goworkers.go#L5)

`type Pool struct { ... }`

Pool is a workers pool

## Examples

```golang
ch := make(chan int)

pool := Init(func() {
    i := <-ch
    fmt.Println(i * i)
})

pool.Add(1)
ch <- 2
ch <- 3
wait := pool.Stop()
ch <- 4
<-wait
```

 Output:

```
4
9
16
```
