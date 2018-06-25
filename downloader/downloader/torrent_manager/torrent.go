package torrent_manager

import (
	"github.com/anacrolix/torrent"
	"log"
	"fmt"
	"os"
	"../tar"
)

type TorrentManager struct {
	urls []string
	client *torrent.Client
}

func New() *TorrentManager {
	manager := &TorrentManager{}

	client, err := torrent.NewClient(nil)
	if err != nil {
		log.Println(err)
	}

	manager.client = client

	defer manager.client.Close()

	return manager
}

func (m *TorrentManager) DownloadTorrent(requestId string, url string) string {
	m.urls = append(m.urls, url)

	t, _ := m.client.AddMagnet(url)

	<-t.GotInfo()
	t.DownloadAll()

	m.client.WaitAll()

	tarName := fmt.Sprintf("%s.tar", requestId)
	tar.TarFolder(tarName, t.Info().Name)

	os.RemoveAll(t.Info().Name)

	log.Printf("Finished %s", t.Info().Name)

	return tarName
}