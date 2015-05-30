package rtorrent

import (
	"crypto/tls"
	"net/http"

	"github.com/kolo/xmlrpc"
)

// RTorrent is used to communicate with a remote rTorrent instance
type RTorrent struct {
	addr         string
	xmlrpcClient *xmlrpc.Client
}

// New returns a new instance of RTorrent
// Pass in a true value for insecure to turn off certificate verification
func New(addr string, insecure bool) (*RTorrent, error) {
	transport := &http.Transport{}
	if insecure {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	newClient, err := xmlrpc.NewClient(addr, transport)
	if err != nil {
		return nil, err
	}

	return &RTorrent{
		addr:         addr,
		xmlrpcClient: newClient,
	}, nil
}

// IP returns the IP reported by this RTorrent instance
func (r *RTorrent) IP() (*string, error) {
	var result string
	err := r.xmlrpcClient.Call("get_ip", nil, &result)
	return &result, err
}

// Name returns the name reported by this RTorrent instance
func (r *RTorrent) Name() (*string, error) {
	var result string
	err := r.xmlrpcClient.Call("get_name", nil, &result)
	return &result, err
}

// DownTotal returns the total downloaded metric reported by this RTorrent instance (bytes)
func (r *RTorrent) DownTotal() (int, error) {
	var result int
	err := r.xmlrpcClient.Call("get_down_total", nil, &result)
	return result, err
}

// UpTotal returns the total uploaded metric reported by this RTorrent instance (bytes)
func (r *RTorrent) UpTotal() (int, error) {
	var result int
	err := r.xmlrpcClient.Call("get_up_total", nil, &result)
	return result, err
}
