package wisdom

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
)

const (
	// HostTypeDNS will fetch DNS names (FQDNs) only
	HostTypeDNS = "dns"

	// HostTypeIP will fetch IPs with TCP protocol and non-zero port number
	HostTypeIP = "ip"
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

func (h *WisdomHosts) Hosts(scope string, size int) ([]string, error) {
	reqURL, err := url.Parse("https://api.open.wisdom.alphasoc.net/v2/items")
	if err != nil {
		return nil, err
	}
	q := reqURL.Query()
	q.Set("category", h.category)
	q.Set("type", h.hostType)
	q.Set("limit", "1000") // the actual limit is much lower, but we want everything
	if scope != "" {
		q.Set("family", scope)
	}
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

	// pick up random hosts
	var hosts []string
	for _, i := range rand.Perm(len(parsed.Items)) {
		if len(hosts) >= size {
			break
		}

		it := parsed.Items[i]

		var host string
		switch h.hostType {
		case HostTypeDNS:
			host = it.Domain
		case HostTypeIP:
			if it.Port <= 0 || it.Protocol != "tcp" {
				continue
			}
			host = net.JoinHostPort(it.IP, strconv.Itoa(it.Port))
		}

		if host != "" {
			hosts = append(hosts, host)
		}
	}

	return hosts, nil
}
