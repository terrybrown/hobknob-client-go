package hobknob

import (
	"strings"
	"time"

	"github.com/coreos/go-etcd/etcd"
)

// Diff contains toggle diffs
type Diff struct {
	name string
	old  bool
	new  bool
}

// Client : the client
type Client struct {
	CacheInterval time.Duration
	AppName       string
	OnUpdate      chan []Diff
	OnError       chan error
	etcd          *etcd.Client
	cache         map[string]bool
	ticker        *time.Ticker
}

// NewClient creates a new instance of the client and returns it
func NewClient(etcdHosts []string, appName string, cacheInterval int) *Client {
	client := &Client{
		cache:         make(map[string]bool),
		etcd:          etcd.NewClient(etcdHosts),
		CacheInterval: time.Duration(cacheInterval) * time.Second,
		AppName:       appName,
		OnUpdate:      make(chan []Diff),
		OnError:       make(chan error),
	}

	return client
}

// Initialise the client
func (c *Client) Initialise() error {
	_, err := c.update()
	if err == nil {
		c.schedule()
	}
	return err
}

func (c *Client) schedule() {
	c.ticker = time.NewTicker(c.CacheInterval)
	go func() {
		for {
			<-c.ticker.C
			diffs, err := c.update()
			if err != nil {
				c.OnError <- err
			}
			if diffs != nil {
				c.OnUpdate <- diffs
			}
		}
	}()
}

func parseValue(val string) (bool, bool) {
	if val == "true" {
		return true, true
	}

	if val == "false" {
		return false, true
	}

	return false, false
}

func parseResponse(resp *etcd.Response) map[string]bool {
	m := make(map[string]bool)
	for _, element := range resp.Node.Nodes {
		ks := strings.Split(element.Key, "/")
		val, ok := parseValue(element.Value)
		if ok {
			m[ks[len(ks)-1]] = val
		}
	}
	return m
}

func diffs(previous map[string]bool, next map[string]bool) []Diff {
	diffs := []Diff{}

	for k, v := range next {
		if v != previous[k] {
			diffs = append(diffs, Diff{
				name: k,
				old:  previous[k],
				new:  v,
			})
		}
	}

	return diffs
}

func (c *Client) update() ([]Diff, error) {
	resp, err := c.etcd.Get("/v1/toggles/"+c.AppName, false, true)
	if err != nil {
		return nil, err
	}

	toggles := parseResponse(resp)
	diffs := diffs(c.cache, toggles)
	c.cache = toggles
	return diffs, nil
}

// Get a toggle state from the cache
func (c *Client) Get(toggle string) (bool, bool) {
	val, ok := c.cache[toggle]
	return val, ok
}

// GetOrDefault get a toggle and supply a default value
func (c *Client) GetOrDefault(toggle string, defaultVal bool) bool {
	if val, ok := c.cache[toggle]; ok {
		return val
	}
	return defaultVal
}
