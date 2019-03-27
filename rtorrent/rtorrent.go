package rtorrent

import (
	"fmt"
	"net/http"

	"github.com/mrobinsn/go-rtorrent/xmlrpc"
	"github.com/pkg/errors"
)

// RTorrent is used to communicate with a remote rTorrent instance
type RTorrent struct {
	addr         string
	xmlrpcClient *xmlrpc.Client
}

// Torrent represents a torrent in rTorrent
type Torrent struct {
	Hash      string
	Name      string
	Path      string
	Size      int
	Label     string
	Completed bool
	Ratio     float64
}

// File represents a file in rTorrent
type File struct {
	Path string
	Size int
}

// View represents a "view" within RTorrent
type View string

const (
	// ViewMain represents the "main" view, containing all torrents
	ViewMain View = "main"
	// ViewStarted represents the "started" view, containing only torrents that have been started
	ViewStarted View = "started"
	// ViewStopped represents the "stopped" view, containing only torrents that have been stopped
	ViewStopped View = "stopped"
	// ViewHashing represents the "hashing" view, containing only torrents that are currently hashing
	ViewHashing View = "hashing"
	// ViewSeeding represents the "seeding" view, containing only torrents that are currently seeding
	ViewSeeding View = "seeding"
)

// Pretty returns a formatted string representing this Torrent
func (t *Torrent) Pretty() string {
	return fmt.Sprintf("Torrent:\n\tHash: %v\n\tName: %v\n\tPath: %v\n\tLabel: %v\n\tSize: %v bytes\n\tCompleted: %v\n\tRatio: %v\n", t.Hash, t.Name, t.Path, t.Label, t.Size, t.Completed, t.Ratio)
}

// Pretty returns a formatted string representing this File
func (f *File) Pretty() string {
	return fmt.Sprintf("File:\n\tPath: %v\n\tSize: %v bytes\n", f.Path, f.Size)
}

// New returns a new instance of `RTorrent`
// Pass in a true value for `insecure` to turn off certificate verification
func New(addr string, insecure bool) *RTorrent {
	return &RTorrent{
		addr:         addr,
		xmlrpcClient: xmlrpc.NewClient(addr, insecure),
	}
}

// WithHTTPClient allows you to a provide a custom http.Client.
func (r *RTorrent) WithHTTPClient(client *http.Client) *RTorrent {
	r.xmlrpcClient = xmlrpc.NewClientWithHTTPClient(r.addr, client)
	return r
}

// Add adds a new torrent by URL
func (r *RTorrent) Add(url string) error {
	_, err := r.xmlrpcClient.Call("load.start", "", url)
	if err != nil {
		return errors.Wrap(err, "load.start XMLRPC call failed")
	}
	return nil
}

// AddTorrent adds a new torrent by the torrent files data
func (r *RTorrent) AddTorrent(data []byte) error {
	_, err := r.xmlrpcClient.Call("load.raw_start", "", data)
	if err != nil {
		return errors.Wrap(err, "load.raw_start XMLRPC call failed")
	}
	return nil
}

// IP returns the IP reported by this RTorrent instance
func (r *RTorrent) IP() (string, error) {
	result, err := r.xmlrpcClient.Call("network.bind_address")
	if err != nil {
		return "", errors.Wrap(err, "network.bind_address XMLRPC call failed")
	}
	if ips, ok := result.([]interface{}); ok {
		result = ips[0]
	}
	if ip, ok := result.(string); ok {
		return ip, nil
	}
	return "", errors.Errorf("result isn't string: %v", result)
}

// Name returns the name reported by this RTorrent instance
func (r *RTorrent) Name() (string, error) {
	result, err := r.xmlrpcClient.Call("system.hostname")
	if err != nil {
		return "", errors.Wrap(err, "system.hostname XMLRPC call failed")
	}
	if names, ok := result.([]interface{}); ok {
		result = names[0]
	}
	if name, ok := result.(string); ok {
		return name, nil
	}
	return "", errors.Errorf("result isn't string: %v", result)
}

// DownTotal returns the total downloaded metric reported by this RTorrent instance (bytes)
func (r *RTorrent) DownTotal() (int, error) {
	result, err := r.xmlrpcClient.Call("throttle.global_down.total")
	if err != nil {
		return 0, errors.Wrap(err, "throttle.global_down.total XMLRPC call failed")
	}
	if totals, ok := result.([]interface{}); ok {
		result = totals[0]
	}
	if total, ok := result.(int); ok {
		return total, nil
	}
	return 0, errors.Errorf("result isn't int: %v", result)
}

// UpTotal returns the total uploaded metric reported by this RTorrent instance (bytes)
func (r *RTorrent) UpTotal() (int, error) {
	result, err := r.xmlrpcClient.Call("throttle.global_up.total")
	if err != nil {
		return 0, errors.Wrap(err, "throttle.global_up.total XMLRPC call failed")
	}
	if totals, ok := result.([]interface{}); ok {
		result = totals[0]
	}
	if total, ok := result.(int); ok {
		return total, nil
	}
	return 0, errors.Errorf("result isn't int: %v", result)
}

// GetTorrents returns all of the torrents reported by this RTorrent instance
func (r *RTorrent) GetTorrents(view View) ([]Torrent, error) {
	args := []interface{}{"", string(view), "d.name=", "d.size_bytes=", "d.hash=", "d.custom1=", "d.base_path=", "d.is_active=", "d.complete=", "d.ratio="}
	results, err := r.xmlrpcClient.Call("d.multicall2", args...)
	var torrents []Torrent
	if err != nil {
		return torrents, errors.Wrap(err, "d.multicall2 XMLRPC call failed")
	}
	for _, outerResult := range results.([]interface{}) {
		for _, innerResult := range outerResult.([]interface{}) {
			torrentData := innerResult.([]interface{})
			torrents = append(torrents, Torrent{
				Hash:      torrentData[2].(string),
				Name:      torrentData[0].(string),
				Path:      torrentData[4].(string),
				Size:      torrentData[1].(int),
				Label:     torrentData[3].(string),
				Completed: torrentData[6].(int) > 0,
				Ratio:     float64(torrentData[7].(int)) / float64(1000),
			})
		}
	}
	return torrents, nil
}

// Delete removes the torent
func (r *RTorrent) Delete(t Torrent) error {
	_, err := r.xmlrpcClient.Call("d.erase", t.Hash)
	if err != nil {
		return errors.Wrap(err, "d.erase XMLRPC call failed")
	}
	return nil
}

// GetFiles returns all of the files for a given `Torrent`
func (r *RTorrent) GetFiles(t Torrent) ([]File, error) {
	args := []interface{}{t.Hash, 0, "f.path=", "f.size_bytes="}
	results, err := r.xmlrpcClient.Call("f.multicall", args...)
	var files []File
	if err != nil {
		return files, errors.Wrap(err, "f.multicall XMLRPC call failed")
	}
	for _, outerResult := range results.([]interface{}) {
		for _, innerResult := range outerResult.([]interface{}) {
			fileData := innerResult.([]interface{})
			files = append(files, File{
				Path: fileData[0].(string),
				Size: fileData[1].(int),
			})
		}
	}
	return files, nil
}

// SetLabel sets the label on the given Torrent
func (r *RTorrent) SetLabel(t Torrent, newLabel string) error {
	t.Label = newLabel
	args := []interface{}{t.Hash, newLabel}
	if _, err := r.xmlrpcClient.Call("d.custom1.set", args...); err != nil {
		return errors.Wrap(err, "d.custom1.set XMLRPC call failed")
	}
	return nil
}
