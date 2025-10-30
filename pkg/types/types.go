package types

type TransmissionRequest struct {
	Method    string                 `json:"method"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type TorrentInfo struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DownloadDir string `json:"downloadDir"`
	HashString  string `json:"hashString"`
}

type TransmissionResponse struct {
	Arguments struct {
		Torrents []TorrentInfo `json:"torrents"`
	} `json:"arguments"`
	Result string `json:"result"`
}

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Dirs     []string
}
