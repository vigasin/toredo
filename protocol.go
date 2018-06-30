package toredo

const (
	MsgDownload = "Download"
	MsgRemove   = "Remove"
	MsgInfo     = "Info"
)

type DownloaderMessage struct {
	RequestId   string
	MessageType string // Download, Remove, Info
	Url         string
}

type TransfererMessage struct {
	RequestId string
	Url       string
}

