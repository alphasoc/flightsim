package wisdom

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type WisdomHosts struct {
	Family string

	category string
	hostType string
}

func NewWisdomHosts(category, hostType string) *WisdomHosts {
	return &WisdomHosts{
		category: category,
		hostType: hostType,
	}
}

func (h *WisdomHosts) Hosts(size int) ([]string, error) {
	reqURL, err := url.Parse("https://api.open.wisdom.alphasoc.net/v2/items")
	if err != nil {
		return nil, err
	}
	q := reqURL.Query()
	q.Set("category", h.category)
	q.Set("type", h.hostType)
	q.Set("limit", strconv.Itoa(size))
	reqURL.RawQuery = q.Encode()

	if h.Family != "" {
		reqURL.Query().Set("family", h.Family)
	}

	resp, err := http.Get(reqURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if c := resp.StatusCode; c != http.StatusOK {
		b, _ := ioutil.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("api.open.wisdom.alphasoc.net said: %d: %s", c, b)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "api.open.wisdom.alphasoc.net read body error")
	}

	var parsed struct {
		Items []struct {
			Domain   string
			IP       string
			Port     int
			Protocol string
		}
	}

	if err := json.Unmarshal(b, &parsed); err != nil {
		return nil, errors.Wrapf(err, "api.open.wisdom.alphasoc.net parse body error")
	}

	var hosts []string
	for _, it := range parsed.Items {
		h := it.Domain
		if h == "" {
			h = it.IP
		}
		if h == "" {
			continue
		}
		if it.Port > 0 && strings.EqualFold(it.Protocol, "tcp") {
			h = net.JoinHostPort(h, strconv.Itoa(it.Port))
		}
		hosts = append(hosts, h)
	}

	return hosts, nil
}
