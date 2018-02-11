package simulator

import (
	"math/rand"
	"net"
	"strings"
	"time"
)

//List of domain from https://github.com/mailcheck/mailcheck/wiki/List-of-Popular-Domains
var domains = []string{
	/* Default domains included */
	"aol.com", "att.net", "comcast.net", "facebook.com", "gmail.com", "gmx.com", "googlemail.com",
	"google.com", "hotmail.com", "hotmail.co.uk", "mac.com", "me.com", "mail.com", "msn.com",
	"live.com", "sbcglobal.net", "verizon.net", "yahoo.com", "yahoo.co.uk",

	/* Other global domains */
	"email.com", "fastmail.fm", "games.com" /* AOL */, "gmx.net", "hush.com", "hushmail.com", "icloud.com",
	"iname.com", "inbox.com", "lavabit.com", "love.com" /* AOL */, "outlook.com", "pobox.com", "protonmail.com",
	"rocketmail.com" /* Yahoo */, "safe-mail.net", "wow.com" /* AOL */, "ygm.com", /* AOL */
	"ymail.com" /* Yahoo */, "zoho.com", "yandex.com",

	/* United States ISP domains */
	"bellsouth.net", "charter.net", "cox.net", "earthlink.net", "juno.com",

	/* British ISP domains */
	"btinternet.com", "virginmedia.com", "blueyonder.co.uk", "live.co.uk",
	"ntlworld.com", "orange.net", "sky.com", "talktalk.co.uk", "tiscali.co.uk",
	"virgin.net", "bt.com",

	/* Domains used in Asia */
	"sina.com", "qq.com", "naver.com", "hanmail.net", "daum.net", "nate.com", "yahoo.co.jp", "yahoo.co.kr", "yahoo.co.id", "yahoo.co.in", "yahoo.com.sg", "yahoo.com.ph",

	/* French ISP domains */
	"hotmail.fr", "live.fr", "laposte.net", "yahoo.fr", "wanadoo.fr", "orange.fr", "gmx.fr", "sfr.fr", "neuf.fr", "free.fr",

	/* German ISP domains */
	"gmx.de", "hotmail.de", "live.de", "online.de", "t-online.de" /* T-Mobile */, "web.de", "yahoo.de",

	/* Italian ISP domains */
	"libero.it", "virgilio.it", "hotmail.it", "aol.it", "tiscali.it", "alice.it", "live.it", "yahoo.it", "email.it", "tin.it", "poste.it", "teletu.it",

	/* Russian ISP domains */
	"mail.ru", "rambler.ru", "yandex.ru", "ya.ru", "list.ru",

	/* Belgian ISP domains */
	"hotmail.be", "live.be", "skynet.be", "voo.be", "tvcablenet.be", "telenet.be",

	/* Argentinian ISP domains */
	"hotmail.com.ar", "live.com.ar", "yahoo.com.ar", "fibertel.com.ar", "speedy.com.ar", "arnet.com.ar",

	/* Domains used in Mexico */
	"yahoo.com.mx", "live.com.mx", "hotmail.es", "prodigy.net.mx",

	/* Domains used in Brazil */
	"yahoo.com.br", "hotmail.com.br", "outlook.com.br", "uol.com.br", "bol.com.br", "terra.com.br", "ig.com.br", "r7.com", "zipmail.com.br", "globo.com", "globomail.com", "oi.com.br",
}

// Spambot simulator.
type Spambot struct{}

// NewSpambot creates port scan simulator.
func NewSpambot() *Spambot {
	return &Spambot{}
}

// Simulate port scanning for given host.
func (*Spambot) Simulate(extIP net.IP, host string) error {
	d := &net.Dialer{
		LocalAddr: &net.TCPAddr{IP: extIP},
		Timeout:   100 * time.Millisecond,
	}

	conn, err := d.Dial("tcp", host)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

// Hosts returns host:port generated from RFC 1918 addresses.
func (s *Spambot) Hosts() ([]string, error) {
	const nLookup = 10
	var (
		hosts []string
		idx   = rand.Perm(len(domains))
	)

	for i, n := 0, 0; i < nLookup && n < len(domains); i, n = i+1, n+1 {
		mx, err := net.LookupMX(domains[idx[n]])
		if err != nil || len(mx) == 0 {
			i--
			continue
		}
		host := strings.TrimSuffix(mx[0].Host, ".")
		hosts = append(hosts, net.JoinHostPort(host, "25"))
	}
	return hosts, nil
}
