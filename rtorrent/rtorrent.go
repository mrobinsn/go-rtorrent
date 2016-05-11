package rtorrent

import (
	"fmt"

	"github.com/tehjojo/go-rtorrent/xmlrpc"
)

// RTorrent is used to communicate with a remote rTorrent instance
type RTorrent struct {
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
		xmlrpcClient: xmlrpc.NewClient(addr, insecure),
	}
}

// SetAuth sets the authentication for communicating with rTorrent
func (r *RTorrent) SetAuth(username string, password string) {
	r.xmlrpcClient.SetBasicAuth(username, password)
}

// Add adds a new torrent by URL
func (r *RTorrent) Add(url string) error {
	_, err := r.xmlrpcClient.Call("load_start", url)
	if err != nil {
		return err
	}
	return nil
}

// AddTorrent adds a new torrent by the torrent files data
func (r *RTorrent) AddTorrent(data []byte) error {
	_, err := r.xmlrpcClient.Call("load_raw_start", data)
	if err != nil {
		return err
	}
	return nil
}

// IP returns the IP reported by this RTorrent instance
func (r *RTorrent) IP() (string, error) {
	result, err := r.xmlrpcClient.Call("get_ip")
	return result.(string), err
}

// Name returns the name reported by this RTorrent instance
func (r *RTorrent) Name() (string, error) {
	result, err := r.xmlrpcClient.Call("get_name")
	return result.(string), err
}

// DownTotal returns the total downloaded metric reported by this RTorrent instance (bytes)
func (r *RTorrent) DownTotal() (int, error) {
	result, err := r.xmlrpcClient.Call("get_down_total")
	return result.(int), err
}

// UpTotal returns the total uploaded metric reported by this RTorrent instance (bytes)
func (r *RTorrent) UpTotal() (int, error) {
	result, err := r.xmlrpcClient.Call("get_up_total")
	return result.(int), err
}

// GetTorrents returns all of the torrents reported by this RTorrent instance
func (r *RTorrent) GetTorrents(view View) ([]Torrent, error) {
	args := []interface{}{string(view), "d.get_name=", "d.get_size_bytes=", "d.get_hash=", "d.get_custom1=", "d.get_base_path=", "d.is_active=", "d.get_complete=", "d.get_ratio="}
	results, err := r.xmlrpcClient.Call("d.multicall", args...)
	var torrents []Torrent
	if err != nil {
		return torrents, err
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

// GetFiles returns all of the files for a given `Torrent`
func (r *RTorrent) GetFiles(t Torrent) ([]File, error) {
	args := []interface{}{t.Hash, 0, "f.get_path=", "f.get_size_bytes="}
	results, err := r.xmlrpcClient.Call("f.multicall", args...)
	var files []File
	if err != nil {
		return files, err
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
	if _, err := r.xmlrpcClient.Call("d.set_custom1", args...); err != nil {
		return err
	}
	return nil
}
