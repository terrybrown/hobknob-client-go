package hobknob

import (
	"fmt"
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

func parseMultiToggleFeature(m map[string]bool, featureNode *etcd.Node) {
	metaDataKey := featureNode.Key + "/@meta"
	for _, toggleNode := range featureNode.Nodes {
		if toggleNode.Key != metaDataKey {
			val, ok := parseValue(toggleNode.Value)
			if ok {
				m[toggleNode.Key] = val
			}
		}
	}
}

func parseResponse(resp *etcd.Response) map[string]bool {
	m := make(map[string]bool)
	for _, element := range resp.Node.Nodes {
		if element.Dir {
			parseMultiToggleFeature(m, element)
		} else {
			val, ok := parseValue(element.Value)
			if ok {
				m[element.Key] = val
			}
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

func getFeatureToggleKey(appName string, feature string, toggle string) string {
	if toggle == "" {
		return fmt.Sprintf("/v1/toggles/%v/%v", appName, feature)
	}
	return fmt.Sprintf("/v1/toggles/%v/%v/%v", appName, feature, toggle)
}

//Get get a simple toggle state from the cache
func (c *Client) Get(feature string) (bool, bool) {
	val, ok := c.cache[getFeatureToggleKey(c.AppName, feature, "")]
	return val, ok
}

//GetMulti get a multi toggle state from the cache
func (c *Client) GetMulti(feature string, toggle string) (bool, bool) {
	val, ok := c.cache[getFeatureToggleKey(c.AppName, feature, toggle)]
	return val, ok
}

//GetOrDefault get a simple toggle and supply a default value
func (c *Client) GetOrDefault(feature string, defaultVal bool) bool {
	if val, ok := c.cache[getFeatureToggleKey(c.AppName, feature, "")]; ok {
		return val
	}
	return defaultVal
}

//GetOrDefaultMulti get a multi toggle and supply a default value
func (c *Client) GetOrDefaultMulti(feature string, toggle string, defaultVal bool) bool {
	if val, ok := c.cache[getFeatureToggleKey(c.AppName, feature, toggle)]; ok {
		return val
	}
	return defaultVal
}
