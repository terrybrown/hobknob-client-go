hobknob-client-go
---

Client for hobknob written in golang

![Build Status](https://travis-ci.org/opentable/hobknob-client-go.png?branch=master)

Installation:

```shell

go get github.com/opentable/hobknob-client-go

```

Usage:

```go

package main

import (
    "fmt"
    "time"

    hobknob "github.com/opentable/hobknob-client-go"
)

func main() {

    c := hobknob.NewClient([]string{"http://127.0.0.1:4001"}, "testApp", 2)

    err := c.Initialise()

    if err != nil {
        fmt.Println(err)
    }

    go func() {
        for {
            diffs := <-c.OnUpdate
            fmt.Printf("update elapsed, diffs: %v\n", diffs)
        }
    }()

    go func() {
        for {
            err := <-c.OnError
            fmt.Printf("error: %v\n", err)
        }
    }()

    ticker := time.NewTicker(time.Millisecond * 500)
    go func() {
        for t := range ticker.C {
            val, _ := c.Get("myFeature")
            fmt.Printf("%v testApp/myFeature: %v\n", t, val)

            val, _ := c.GetMulti("domainFeature", "com")
            fmt.Printf("%v testApp/domainFeature/com: %v\n", t, val)
        }
    }()

    time.Sleep(time.Second * 15)
    ticker.Stop()
}


```

Docs:

### hobknob.NewClient(etcdHosts []string, appName string, cacheInterval int)
creates a new instance of the hobknob-client

params:
- etcdHosts // array of hosts
- appName // identifier for this app to find its toggles
- cacheInterval // number of seconds between updates

returns a new client

### client.Initialise()
run once to initialise the client you just created
returns an error or nil

### client.Get(feature string)
returns (value bool, exists bool) for the simple feature toggle `appName/feature`

### client.GetMulti(feature string, toggle string)
returns (value bool, exists bool) for the multi feature toggle `appName/feature/toggle`

### client.GetOrDefault(feature string, defaultVal bool)
returns the value of the simple feature toggle `appName/feature`, or the supplied default if it didn't exist

### client.GetOrDefaultMulti(feature string, toggle string, defaultVal bool)
returns the value of the multi feature toggle `appName/feature/toggle`, or the supplied default if it didn't exist

### client.OnUpdate chan []Diff
a channel which gets published every time an update happens. Contains an array of Diffs for toggles that changed in the last update

```go
Diff {
  name string
  old bool
  new bool
}
```

### client.OnError chan error
a channel for reading errors which might occur.
