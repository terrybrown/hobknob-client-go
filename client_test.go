package hobknob

import (
	"testing"
	"time"
)

func Setup() (*Client, error) {
	c := NewClient([]string{"http://127.0.0.1:4001"}, "testApp", 1)
	err := c.Initialise()

	return c, err
}

func TestNew(t *testing.T) {
	c, _ := Setup()

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
	_, err := Setup()

	if err != nil {
		t.Error(err)
	}
}

func TestGet(t *testing.T) {
	c, err := Setup()

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
	c, err := Setup()

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

func TestGetOrDefault(t *testing.T) {
	c, err := Setup()

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

func TestSchedule(t *testing.T) {
	c, err := Setup()

	if err != nil {
		t.Error(err)
	}

	updateError := <-c.OnUpdate

	if updateError != nil {
		t.Fatalf("Got an error when updating %v", updateError)
	}
}

func BenchmarkGet(b *testing.B) {
	c, err := Setup()

	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		c.Get("mytoggle")
	}
}

func BenchmarkGetOrDefault(b *testing.B) {
	c, err := Setup()

	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		c.GetOrDefault("mytoggle", true)
	}
}
