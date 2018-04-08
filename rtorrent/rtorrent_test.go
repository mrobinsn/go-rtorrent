package rtorrent

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestRTorrent(t *testing.T) {
	/*
		These tests rely on a local instance of rtorrent to be running in a clean state.
		Use the included `test.sh` script to run these tests.
	*/
	client := New("http://localhost/RPC2", false)

	t.Run("get ip", func(t *testing.T) {
		ip, err := client.IP()
		require.NoError(t, err)
		require.NotEmpty(t, ip)
		require.Equal(t, "0.0.0.0", ip)
	})

	t.Run("get name", func(t *testing.T) {
		name, err := client.Name()
		require.NoError(t, err)
		require.NotEmpty(t, name)
	})

	t.Run("down total", func(t *testing.T) {
		total, err := client.DownTotal()
		require.NoError(t, err)
		require.Zero(t, total, "expected no data to be transferred yet")
	})

	t.Run("up total", func(t *testing.T) {
		total, err := client.UpTotal()
		require.NoError(t, err)
		require.Zero(t, total, "expected no data to be transferred yet")
	})

	t.Run("get no torrents", func(t *testing.T) {
		torrents, err := client.GetTorrents(ViewMain)
		require.NoError(t, err)
		require.Empty(t, torrents, "expected no torrents to be added yet")
	})

	t.Run("add", func(t *testing.T) {
		t.Run("by url", func(t *testing.T) {
			err := client.Add("https://archive.org/download/bergeysmanualofd1957amer/bergeysmanualofd1957amer_archive.torrent")
			require.NoError(t, err)

			t.Run("get torrent", func(t *testing.T) {
				// It will take some time to appear, so retry a few times
				tries := 0
				var torrents []Torrent
				var err error
				for {
					<-time.After(time.Second)
					torrents, err = client.GetTorrents(ViewMain)
					require.NoError(t, err)
					if len(torrents) > 0 {
						break
					}
					if tries > 10 {
						require.NoError(t, errors.Errorf("torrent did not show up in time"))
					}
					tries++
				}
				require.NotEmpty(t, torrents)
				require.Len(t, torrents, 1)
				require.Equal(t, "A3302641AF31E17B4DCCDDDC2B41E208E85D03A2", torrents[0].Hash)
				require.Equal(t, "bergeysmanualofd1957amer", torrents[0].Name)
				require.Equal(t, "", torrents[0].Label)
				require.Equal(t, 888773112, torrents[0].Size)
				require.Equal(t, "/downloads/incoming/bergeysmanualofd1957amer", torrents[0].Path)
				require.False(t, torrents[0].Completed)

				t.Run("get files", func(t *testing.T) {
					files, err := client.GetFiles(torrents[0])
					require.NoError(t, err)
					require.NotEmpty(t, files)
					require.Len(t, files, 19)
					for _, f := range files {
						require.NotEmpty(t, f.Path)
						require.NotZero(t, f.Size)
					}
				})

				t.Run("change label", func(t *testing.T) {
					err := client.SetLabel(torrents[0], "TestLabel")
					require.NoError(t, err)

					// It will take some time to change, so try a few times
					tries := 0
					for {
						<-time.After(time.Second)
						torrents, err = client.GetTorrents(ViewMain)
						require.NoError(t, err)
						require.Len(t, torrents, 1)
						if torrents[0].Label != "" {
							break
						}
						if tries > 10 {
							require.NoError(t, errors.Errorf("torrent label did not change in time"))
						}
						tries++
					}
					require.Equal(t, "TestLabel", torrents[0].Label)
				})

				t.Run("delete torrent", func(t *testing.T) {
					err := client.Delete(torrents[0])
					require.NoError(t, err)

					torrents, err := client.GetTorrents(ViewMain)
					require.NoError(t, err)
					require.Empty(t, torrents)

					t.Run("get torrent", func(t *testing.T) {
						// It will take some time to disappear, so retry a few times
						tries := 0
						var torrents []Torrent
						var err error
						for {
							<-time.After(time.Second)
							torrents, err = client.GetTorrents(ViewMain)
							require.NoError(t, err)
							if len(torrents) == 0 {
								break
							}
							if tries > 10 {
								require.NoError(t, errors.Errorf("torrent did not delete in time"))
							}
							tries++
						}
						require.Empty(t, torrents)
					})
				})
			})
		})

		t.Run("with data", func(t *testing.T) {
			b, err := ioutil.ReadFile("testdata/bergeysmanualofd1957amer_archive.torrent")
			require.NoError(t, err)
			require.NotEmpty(t, b)

			err = client.AddTorrent(b)
			require.NoError(t, err)

			t.Run("get torrent", func(t *testing.T) {
				// It will take some time to appear, so retry a few times
				tries := 0
				var torrents []Torrent
				var err error
				for {
					<-time.After(time.Second)
					torrents, err = client.GetTorrents(ViewMain)
					require.NoError(t, err)
					if len(torrents) > 0 {
						break
					}
					if tries > 10 {
						require.NoError(t, errors.Errorf("torrent did not show up in time"))
					}
					tries++
				}
				require.NotEmpty(t, torrents)
				require.Len(t, torrents, 1)
				require.Equal(t, "A3302641AF31E17B4DCCDDDC2B41E208E85D03A2", torrents[0].Hash)
				require.Equal(t, "bergeysmanualofd1957amer", torrents[0].Name)
				require.Equal(t, "", torrents[0].Label)
				require.Equal(t, 888773112, torrents[0].Size)
				require.Equal(t, "/downloads/incoming/bergeysmanualofd1957amer", torrents[0].Path)
				require.False(t, torrents[0].Completed)

				t.Run("get files", func(t *testing.T) {
					files, err := client.GetFiles(torrents[0])
					require.NoError(t, err)
					require.NotEmpty(t, files)
					require.Len(t, files, 19)
					for _, f := range files {
						require.NotEmpty(t, f.Path)
						require.NotZero(t, f.Size)
					}
				})

				t.Run("delete torrent", func(t *testing.T) {
					err := client.Delete(torrents[0])
					require.NoError(t, err)

					torrents, err := client.GetTorrents(ViewMain)
					require.NoError(t, err)
					require.Empty(t, torrents)

					t.Run("get torrent", func(t *testing.T) {
						// It will take some time to disappear, so retry a few times
						tries := 0
						var torrents []Torrent
						var err error
						for {
							<-time.After(time.Second)
							torrents, err = client.GetTorrents(ViewMain)
							require.NoError(t, err)
							if len(torrents) == 0 {
								break
							}
							if tries > 10 {
								require.NoError(t, errors.Errorf("torrent did not delete in time"))
							}
							tries++
						}
						require.Empty(t, torrents)
					})
				})
			})
		})
	})

	t.Run("down total post activity", func(t *testing.T) {
		total, err := client.DownTotal()
		require.NoError(t, err)
		require.NotZero(t, total, "expected data to be transferred")
	})

	t.Run("up total post activity", func(t *testing.T) {
		total, err := client.UpTotal()
		require.NoError(t, err)
		require.NotZero(t, total, "expected data to be transferred")
	})
}
