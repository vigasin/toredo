package toredo

type ApiRequest struct {
	Url string
}

type ApiResponse struct {
	RequestId string
}

const (
	MsgDownloaderInDownload = "Download"
	MsgDownloaderInRemove   = "Remove"
	MsgDownloaderInInfo     = "Info"
)

const (
	MsgDownloaderOutDownloaded = "Downloaded"
	MsgDownloaderOutRemoved    = "Removed"
	MsgDownloaderOutGotInfo    = "GotInfo"
)

type DownloaderInMessage struct {
	MessageType string // Download, Remove, Info
	RequestId   string

	Url string
}

type DownloaderOutMessage struct {
	MessageType string // Downloaded, Removed, GotInfo
	RequestId   string

	Url     string
	Message string
}

type TransfererMessage struct {
	RequestId string
	Url       string
}
