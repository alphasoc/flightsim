package simulator

import (
	"context"
	"math/rand"
	"net"
	"net/http"

	"github.com/cretz/bine/tor"
)

//Slice containing websites hosted by torproject
var torHosts = []string{"expyuzz4wqqyqhjn.onion", "qrmfuxwgyzk5jdjz.onion", "e4nybovdbcwaqlyt.onion", "52g5y5karruvc7bz.onion", "x3nelbld33llasqv.onion", "vijs2fmpd72nbqok.onion",
	"z5tfsnikzulwicxs.onion", "icxe4yp32mq6gm6n.onion", "qigcb4g4xxbh5ho6.onion", "kkvj4mhsttfcrksj.onion", "3gldbgtv5e4god56.onion", "tgnv2pssfumdedyw.onion",
	"5bam5t36aombgv76.onion", "sdscoq7snqtznauu.onion", "rqef5a5mebgq46y5.onion", "ruv6ue7d3t22el2a.onion", "zfu7x4fuagirknhb.onion", "klbl4glo2btuwyok.onion",
	"ngp5wfw5z6ms3ynx.onion", "tngjm3owsslo3wgo.onion", "dccbbv6cooddgcrq.onion", "jqs44zhtxl2uo6gk.onion", "odz6noxeukaw43e7.onion", "54nujbl4qohb5qdp.onion",
	"eibwzyiqgk6vgugg.onion", "f7lqb5oicvsahone.onion", "y7pm6of53hzeb7u2.onion", "n46o4uxsej2icp5l.onion", "rougmnvswfsmd4dq.onion", "l3xrunzkfufzvw2c.onion",
	"kzcx36ytbsm5iogs.onion", "ebxqgaz3dwywcoxl.onion", "yz7lpwfhhzcdyc5y.onion", "tgel7v4rpcllsrk2.onion", "llhb3u5h3q66ha62.onion", "rh7jaux2r3tzrqp4.onion",
	"sbe5fi5cka5l3fqe.onion", "koz2sqqf4w23qxw2.onion", "hyntj47ow4ermsrh.onion", "yabd3wlpvybdnvzg.onion", "c5qrls2slxqz6vdw.onion", "wcgqzqyfi7a6iu62.onion",
	"6m6blys5mwg2jwex.onion", "fhny6b7b6sbslc2b.onion", "s2bweojt5vg52e5i.onion", "xlv5dckljs4vhmhm.onion", "lfdhmyq24uacliu5.onion", "vt5hknv6sblkgf22.onion",
	"buqlpzbbcyat2jiy.onion", "bn6kma5cpxill4pe.onion", "4bflp2c4tnynnbes.onion", "2xcd24wfjiqwzwnr.onion", "dgvdmophvhunawds.onion", "fylvgu5r6gcdadeo.onion",
	"2iqyjmvrkrq5h5mg.onion", "nraswjtnyrvywxk7.onion", "ea5faa5po25cf7fb.onion", "krkzagd5yo4bvypt.onion", "hzmun3rnnxjhkyhg.onion", "expyuzz4wqqyqhjn.onion",
}

type TorSimulator struct {
	tor     *tor.Tor
	initerr error
}

//Returns new TorSimulator
func NewTorSimulator() *TorSimulator {
	tor, err := tor.Start(nil, &tor.StartConf{RetainTempDataDir: false})
	tor.StopProcessOnClose = true
	return &TorSimulator{tor: tor, initerr: err}
}

//Returns random hosts from the slice limited by the parameter "size"
func (t TorSimulator) Hosts(scope string, size int) ([]string, error) {
	var hosts []string
	for _, i := range rand.Perm(len(torHosts)) {
		if len(hosts) >= size {
			break
		}
		hosts = append(hosts, torHosts[i])
	}
	return hosts, nil
}

//Simulates connection to tor network
func (t TorSimulator) Simulate(ctx context.Context, bind net.IP, dst string) error {
	if t.initerr != nil {
		panic(t.initerr)
	}
	//TODO close tor connection

	dialer, err := t.tor.Dialer(ctx, nil)
	if err != nil {
		return err
	}

	httpClient := &http.Client{Transport: &http.Transport{DialContext: dialer.DialContext}}
	resp, err := httpClient.Get("http://" + dst)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
