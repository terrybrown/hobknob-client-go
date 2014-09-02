hobknob-client-go
---

Client for hobknob written in golang

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

	hobknob "github.com/opentable/go-hobknob"
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
			val, _ := c.Get("mytoggle")
			fmt.Printf("%v testApp/mytoggle: %v\n", t, val)
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

### client.Get(toggle string)
returns (value bool, exists bool)

### client.GetOrDefault(toggle string, defaultVal bool)
returns the value of the toggle, or the supplied default if it didn't exist

### client.OnUpdate chan string
a channel which gets published every time an update happens. Contains an array of Diffs for toggles that changed in the last update

Diff {
  name string
  old bool
  new bool
}

### client.OnError chan error
a channel for reading errors which might occur.
