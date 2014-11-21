package hobknob

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func readTestResponseFromJsonFile() string {
	fileBytes, err := ioutil.ReadFile("testEtcdResponse.json")
	if err != nil {
		panic(err)
	}
	return string(fileBytes)
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	fmt.Fprintf(w, readTestResponseFromJsonFile())
}

func initServer() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":4001", nil)
}

func TestServer(t *testing.T) {
	go initServer()
}

func Setup(t *testing.T) (*Client, error) {
	c := NewClient([]string{"http://127.0.0.1:4001"}, "testApp", 1)
	err := c.Initialise()

	go func() {
		for {
			err := <-c.OnError
			t.Error(err)
		}
	}()

	go func() {
		for {
			<-c.OnUpdate
		}
	}()

	return c, err
}

func SetupBench(b *testing.B) (*Client, error) {
	c := NewClient([]string{"http://127.0.0.1:4001"}, "testApp", 1)
	err := c.Initialise()

	go func() {
		for {
			err := <-c.OnError
			b.Error(err)
		}
	}()

	go func() {
		for {
			<-c.OnUpdate
		}
	}()

	return c, err
}

func TestNew(t *testing.T) {
	c, _ := Setup(t)

	if c == nil {
		t.Fatalf("client was null")
	}

	if c.AppName != "testApp" {
		t.Fatalf("AppName not initialised: %v", c.AppName)
	}

	if c.CacheInterval != (time.Duration(1) * time.Second) {
		t.Fatalf("CacheInterval not initialised %v", c.CacheInterval)
	}
}

func TestInitialise(t *testing.T) {
	_, err := Setup(t)

	if err != nil {
		t.Error(err)
	}
}

func TestGet(t *testing.T) {
	c, err := Setup(t)

	if err != nil {
		t.Error(err)
	}

	toggle, exists := c.Get("mytoggle")

	if toggle != true {
		t.Fatalf("expecting toggle 'mytoggle' to have value 'true' actual: '%v'", toggle)
	}

	if exists != true {
		t.Fatalf("expecting exists 'mytoggle' to have value 'true' actual: '%v'", toggle)
	}
}

func TestGetNonExistentToggle(t *testing.T) {
	c, err := Setup(t)

	if err != nil {
		t.Error(err)
	}

	toggle, exists := c.Get("unknowntoggle")

	if toggle != false {
		t.Fatalf("expecting toggle 'unknowntoggle' to have value 'false' actual: '%v'", toggle)
	}

	if exists != false {
		t.Fatalf("expecting exists 'unknowntoggle' to have value 'false' actual: '%v'", exists)
	}
}

func TestGetBadToggle(t *testing.T) {
	c, err := Setup(t)

	if err != nil {
		t.Error(err)
	}

	toggle, exists := c.Get("badtoggle")

	if toggle != false {
		t.Fatalf("expecting toggle 'badtoggle' to have value 'false' actual: '%v'", toggle)
	}

	if exists != false {
		t.Fatalf("expecting exists 'badtoggle' to have value 'false' actual: '%v'", exists)
	}
}

func TestGetOrDefault(t *testing.T) {
	c, err := Setup(t)

	if err != nil {
		t.Error(err)
	}

	toggle1 := c.GetOrDefault("mytoggle", true)

	if toggle1 != true {
		t.Fatalf("expecting toggle 'mytoggle' to have value 'true' actual: '%v'", toggle1)
	}

	toggle2 := c.GetOrDefault("unknowntoggle", true)

	if toggle2 != true {
		t.Fatalf("expecting toggle 'unknowntoggle' to have value 'true' actual: '%v'", toggle2)
	}
}

func TestGetMulti(t *testing.T) {
	c, err := Setup(t)

	if err != nil {
		t.Error(err)
	}

	toggle, exists := c.GetMulti("multi", "toggle1")

	if toggle != true {
		t.Fatalf("expecting toggle 'multi/toggle1' to have value 'true' actual: '%v'", toggle)
	}

	if exists != true {
		t.Fatalf("expecting exists 'multi/toggle1' to have value 'true' actual: '%v'", toggle)
	}
}

func TestGetMultiWhenFeatureDoesNotExist(t *testing.T) {
	c, err := Setup(t)

	if err != nil {
		t.Error(err)
	}

	toggle, exists := c.GetMulti("unkknownFeature", "toggle1")

	if toggle != false {
		t.Fatalf("expecting toggle 'unknownFeature/toggle1' to have value 'false' actual: '%v'", toggle)
	}

	if exists != false {
		t.Fatalf("expecting exists 'unknownFeature/toggle1' to have value 'false' actual: '%v'", exists)
	}
}

func TestGetMultiWhenFeatureToggleDoesNotExist(t *testing.T) {
	c, err := Setup(t)

	if err != nil {
		t.Error(err)
	}

	toggle, exists := c.GetMulti("multi", "unknownToggle")

	if toggle != false {
		t.Fatalf("expecting toggle 'multi/unknowntoggle' to have value 'false' actual: '%v'", toggle)
	}

	if exists != false {
		t.Fatalf("expecting exists 'multi/unknowntoggle' to have value 'false' actual: '%v'", exists)
	}
}
func TestGetMultiBadToggle(t *testing.T) {
	c, err := Setup(t)

	if err != nil {
		t.Error(err)
	}

	toggle, exists := c.GetMulti("multi", "badtoggle")

	if toggle != false {
		t.Fatalf("expecting toggle 'multi/badtoggle' to have value 'false' actual: '%v'", toggle)
	}

	if exists != false {
		t.Fatalf("expecting exists 'multi/badtoggle' to have value 'false' actual: '%v'", exists)
	}
}

func TestSchedule(t *testing.T) {
	c, err := Setup(t)

	if err != nil {
		t.Error(err)
	}

	diffs := <-c.OnUpdate

	if diffs == nil {
		t.Fatalf("Got a nil update value: %v, was expecting: []Diffs{}", diffs)
	}
}

func BenchmarkGet(b *testing.B) {
	c, err := SetupBench(b)

	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		c.Get("mytoggle")
	}
}

func BenchmarkGetOrDefault(b *testing.B) {
	c, err := SetupBench(b)

	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		c.GetOrDefault("mytoggle", true)
	}
}
