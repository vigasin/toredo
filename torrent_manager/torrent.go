package torrent_manager

import (
	"github.com/vigasin/toredo/tar"
	"fmt"
	"github.com/anacrolix/torrent"
	"log"
	"os"
	"path"
	"io"
)

type TorrentManager struct {
	downloadPath string
	publicPath string
	urls   []string
	client *torrent.Client
}

func New(downloadPath string, publicPath string) *TorrentManager {
	m := &TorrentManager{}

	m.downloadPath = downloadPath
	m.publicPath = publicPath

	config := torrent.NewDefaultClientConfig()
	config.DataDir = downloadPath

	client, err := torrent.NewClient(config)
	if err != nil {
		log.Println(err)
	}

	m.client = client

	return m
}

func (m *TorrentManager) DownloadTorrent(requestId string, url string) string {
	m.urls = append(m.urls, url)

	t, err := m.client.AddMagnet(url)

	if err != nil {
		log.Printf("Error adding magnet: %s\n", err)
	}

	<-t.GotInfo()
	t.DownloadAll()

	// TODO: Should wait for just this torrent, not all
	m.client.WaitAll()

	tarName := fmt.Sprintf("%s.tar", requestId)
	tarPath := path.Join(m.publicPath, tarName)
	tar.TarFolder(tarPath, m.downloadPath, t.Info().Name)

	os.RemoveAll(t.Info().Name)

	log.Printf("Finished %s", t.Info().Name)

	return tarName
}

func (m *TorrentManager) WriteStatus(w io.Writer) {
	m.client.WriteStatus(w)
}
