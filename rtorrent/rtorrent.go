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

// FieldValue contains the Field and Value of an attribute on a rTorrent
type FieldValue struct {
	Field Field
	Value string
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

// Status represents the status of a torrent
type Status struct {
	Completed      bool
	CompletedBytes int
	DownRate       int
	UpRate         int
	Ratio          float64
	Size           int
}

// File represents a file in rTorrent
type File struct {
	Path string
	Size int
}

// Field represents a attribute on a RTorrent entity that can be queried or set
type Field string

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

	// DName represents the name of a "Downloading Items"
	DName Field = "d.name"
	// DLabel represents the label of a "Downloading Item"
	DLabel Field = "d.custom1"
	// DSizeInBytes represents the size in bytes of a "Downloading Item"
	DSizeInBytes Field = "d.syze_bytes"
	// DHash represents the hash of a "Downloading Item"
	DHash Field = "d.hash"
	// DBasePath represents the base path of a "Downloading Item"
	DBasePath Field = "d.base_path"
	// DIsActive represents whether a "Downloading Item" is active or not
	DIsActive Field = "d.is_active"
	// DRatio represents the ratio of a "Downloading Item"
	DRatio Field = "d.ratio"
	// DComplete represents whether the "Downloading Item" is complete or not
	DComplete Field = "d.complete"
	// DCompletedBytes represents the total of completed bytes of the "Downloading Item"
	DCompletedBytes Field = "d.completed_bytes"
	// DDownRate represents the download rate of the "Downloading Item"
	DDownRate Field = "d.down.rate"
	// DUpRate represents the upload rate of the "Downloading Item"
	DUpRate Field = "d.up.rate"

	// FPath represents the path of a "File Item"
	FPath Field = "f.path"
	// FSizeInBytes represents the size in bytes of a "File Item"
	FSizeInBytes Field = "f.size_bytes"
)

// Query converts the field to a string which allows it to be queried
// Example:
//  DName.Query() // returns "d.name="
func (f Field) Query() string {
	return fmt.Sprintf("%s=", f)
}

// SetValue returns a FieldValue struct which can be used to set the field on a particular item in rTorrent to the specified value
func (f Field) SetValue(value string) *FieldValue {
	return &FieldValue{f, value}
}

// Cmd returns the representation of the field which allows it to be used a command with RTorrent
func (f Field) Cmd() string {
	return string(f)
}

func (f *FieldValue) String() string {
	return fmt.Sprintf("%s.set=\"%s\"", f.Field, f.Value)
}

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

// AddStopped adds a new torrent by URL in a stopped state
//
// extraArgs can be any valid rTorrent rpc command. For instance:
//
// Adds the Torrent by URL (stopped) and sets the label on the torrent
//  AddStopped("some-url", &FieldValue{"d.custom1", "my-label"})
// Or:
//  AddStopped("some-url", DLabel.SetValue("my-label"))
//
// Adds the Torrent by URL (stopped) and  sets the label and base path
//  AddStopped("some-url", &FieldValue{"d.custom1", "my-label"}, &FiedValue{"d.base_path", "/some/valid/path"})
// Or:
//  AddStopped("some-url", DLabel.SetValue("my-label"), DBasePath.SetValue("/some/valid/path"))
func (r *RTorrent) AddStopped(url string, extraArgs ...*FieldValue) error {
	return r.add("load.normal", []byte(url), extraArgs...)
}

// Add adds a new torrent by URL and starts the torrent
//
// extraArgs can be any valid rTorrent rpc command. For instance:
//
// Adds the Torrent by URL and sets the label on the torrent
//  Add("some-url", "d.custom1.set=\"my-label\"")
// Or:
//  Add("some-url", DLabel.SetValue("my-label"))
//
// Adds the Torrent by URL and  sets the label as well as base path
//  Add("some-url", "d.custom1.set=\"my-label\"", "d.base_path=\"/some/valid/path\"")
// Or:
//  Add("some-url", DLabel.SetValue("my-label"), DBasePath.SetValue("/some/valid/path"))
func (r *RTorrent) Add(url string, extraArgs ...*FieldValue) error {
	return r.add("load.start", []byte(url), extraArgs...)
}

// AddTorrentStopped adds a new torrent by the torrent files data but does not start the torrent
//
// extraArgs can be any valid rTorrent rpc command. For instance:
//
// Adds the Torrent file (stopped) and sets the label on the torrent
//  AddTorrentStopped(fileData, "d.custom1.set=\"my-label\"")
// Or:
//  AddTorrentStopped(fileData, DLabel.SetValue("my-label"))
//
// Adds the Torrent file and (stopped) sets the label and base path
//  AddTorrentStopped(fileData, "d.custom1.set=\"my-label\"", "d.base_path=\"/some/valid/path\"")
// Or:
//  AddTorrentStopped(fileData, DLabel.SetValue("my-label"), DBasePath.SetValue("/some/valid/path"))
func (r *RTorrent) AddTorrentStopped(data []byte, extraArgs ...*FieldValue) error {
	return r.add("load.raw", data, extraArgs...)
}

// AddTorrent adds a new torrent by the torrent files data and starts the torrent
//
// extraArgs can be any valid rTorrent rpc command. For instance:
//
// Adds the Torrent file and sets the label on the torrent
//  Add(fileData, "d.custom1.set=\"my-label\"")
// Or:
//  AddTorrent(fileData, DLabel.SetValue("my-label"))
//
// Adds the Torrent file and  sets the label and base path
//  Add(fileData, "d.custom1.set=\"my-label\"", "d.base_path=\"/some/valid/path\"")
// Or:
//  AddTorrent(fileData, DLabel.SetValue("my-label"), DBasePath.SetValue("/some/valid/path"))
func (r *RTorrent) AddTorrent(data []byte, extraArgs ...*FieldValue) error {
	return r.add("load.raw_start", data, extraArgs...)
}

func (r *RTorrent) add(cmd string, data []byte, extraArgs ...*FieldValue) error {
	args := []interface{}{data}
	for _, v := range extraArgs {
		args = append(args, v.String())
	}

	_, err := r.xmlrpcClient.Call(cmd, "", args)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("%s XMLRPC call failed", cmd))
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

// DownRate returns the current download rate reported by this RTorrent instance (bytes/s)
func (r *RTorrent) DownRate() (int, error) {
	result, err := r.xmlrpcClient.Call("throttle.global_down.rate")
	if err != nil {
		return 0, errors.Wrap(err, "throttle.global_down.rate XMLRPC call failed")
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

// UpRate returns the current upload rate reported by this RTorrent instance (bytes/s)
func (r *RTorrent) UpRate() (int, error) {
	result, err := r.xmlrpcClient.Call("throttle.global_up.rate")
	if err != nil {
		return 0, errors.Wrap(err, "throttle.global_up.rate XMLRPC call failed")
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
	args := []interface{}{"", string(view), DName.Query(), DSizeInBytes.Query(), DHash.Query(), DLabel.Query(), DBasePath.Query(), DIsActive.Query(), DComplete.Query(), DRatio.Query()}
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

// GetTorrent returns the torrent identified by the given hash
func (r *RTorrent) GetTorrent(hash string) (Torrent, error) {
	var t Torrent
	t.Hash = hash
	// Name
	results, err := r.xmlrpcClient.Call("d.name", t.Hash)
	if err != nil {
		return t, errors.Wrap(err, "d.name XMLRPC call failed")
	}
	t.Name = results.([]interface{})[0].(string)
	// Size
	results, err = r.xmlrpcClient.Call("d.size_bytes", t.Hash)
	if err != nil {
		return t, errors.Wrap(err, "d.size_bytes XMLRPC call failed")
	}
	t.Size = results.([]interface{})[0].(int)
	// Label
	results, err = r.xmlrpcClient.Call("d.custom1", t.Hash)
	if err != nil {
		return t, errors.Wrap(err, "d.custom1 XMLRPC call failed")
	}
	t.Label = results.([]interface{})[0].(string)
	// Path
	results, err = r.xmlrpcClient.Call("d.base_path", t.Hash)
	if err != nil {
		return t, errors.Wrap(err, "d.base_path XMLRPC call failed")
	}
	t.Path = results.([]interface{})[0].(string)
	// Completed
	results, err = r.xmlrpcClient.Call("d.complete", t.Hash)
	if err != nil {
		return t, errors.Wrap(err, "d.complete XMLRPC call failed")
	}
	t.Completed = results.([]interface{})[0].(int) > 0
	// Ratio
	results, err = r.xmlrpcClient.Call("d.ratio", t.Hash)
	if err != nil {
		return t, errors.Wrap(err, "d.ratio XMLRPC call failed")
	}
	t.Ratio = float64(results.([]interface{})[0].(int)) / float64(1000)
	return t, nil
}

// Delete removes the torrent
func (r *RTorrent) Delete(t Torrent) error {
	_, err := r.xmlrpcClient.Call("d.erase", t.Hash)
	if err != nil {
		return errors.Wrap(err, "d.erase XMLRPC call failed")
	}
	return nil
}

// GetFiles returns all of the files for a given `Torrent`
func (r *RTorrent) GetFiles(t Torrent) ([]File, error) {
	args := []interface{}{t.Hash, 0, FPath.Query(), FSizeInBytes.Query()}
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

// GetStatus returns the Status for a given Torrent
func (r *RTorrent) GetStatus(t Torrent) (Status, error) {
	var s Status
	// Completed
	results, err := r.xmlrpcClient.Call("d.complete", t.Hash)
	if err != nil {
		return s, errors.Wrap(err, "d.complete XMLRPC call failed")
	}
	s.Completed = results.([]interface{})[0].(int) > 0
	// CompletedBytes
	results, err = r.xmlrpcClient.Call("d.completed_bytes", t.Hash)
	if err != nil {
		return s, errors.Wrap(err, "d.completed_bytes XMLRPC call failed")
	}
	s.CompletedBytes = results.([]interface{})[0].(int)
	// DownRate
	results, err = r.xmlrpcClient.Call("d.down.rate", t.Hash)
	if err != nil {
		return s, errors.Wrap(err, "d.down.rate XMLRPC call failed")
	}
	s.DownRate = results.([]interface{})[0].(int)
	// UpRate
	results, err = r.xmlrpcClient.Call("d.up.rate", t.Hash)
	if err != nil {
		return s, errors.Wrap(err, "d.up.rate XMLRPC call failed")
	}
	s.UpRate = results.([]interface{})[0].(int)
	// Ratio
	results, err = r.xmlrpcClient.Call("d.ratio", t.Hash)
	if err != nil {
		return s, errors.Wrap(err, "d.ratio XMLRPC call failed")
	}
	s.Ratio = float64(results.([]interface{})[0].(int)) / float64(1000)
	// Size
	results, err = r.xmlrpcClient.Call("d.size_bytes", t.Hash)
	if err != nil {
		return s, errors.Wrap(err, "d.size_bytes XMLRPC call failed")
	}
	s.Size = results.([]interface{})[0].(int)
	return s, nil
}
