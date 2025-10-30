package types

type TransmissionRequest struct {
	Method    string                 `json:"method"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type TorrentInfo struct {
	ID             int     `json:"id"`
	Name           string  `json:"name"`
	DownloadDir    string  `json:"downloadDir"`
	HashString     string  `json:"hashString"`
	TotalSize      int64   `json:"totalSize"`
	SizeWhenDone   int64   `json:"sizeWhenDone"`
	LeftUntilDone  int64   `json:"leftUntilDone"`
	RateDownload   int     `json:"rateDownload"`
	RateUpload     int     `json:"rateUpload"`
	PercentDone    float64 `json:"percentDone"`
	Status         int     `json:"status"`
	AddedDate      int64   `json:"addedDate"`
	DoneDate       int64   `json:"doneDate"`
	UploadedEver   int64   `json:"uploadedEver"`
	DownloadedEver int64   `json:"downloadedEver"`
	Ratio          float64 `json:"uploadRatio"`
}

type TransmissionResponse struct {
	Arguments struct {
		Torrents []TorrentInfo `json:"torrents"`
	} `json:"arguments"`
	Result string `json:"result"`
}

// SessionInfo contains Transmission session information
type SessionInfo struct {
	DownloadDir      string  `json:"download-dir"`
	DownloadDirFree  int64   `json:"download-dir-free"`
	PeerPort         int     `json:"peer-port"`
	SeedRatioLimit   float64 `json:"seedRatioLimit"`
	SeedRatioLimited bool    `json:"seedRatioLimited"`
	UploadSpeed      int64   `json:"uploadSpeed"`
	DownloadSpeed    int64   `json:"downloadSpeed"`
	AltSpeedEnabled  bool    `json:"alt-speed-enabled"`
	AltSpeedUp       int     `json:"alt-speed-up"`
	AltSpeedDown     int     `json:"alt-speed-down"`
}

// SessionStats contains Transmission session statistics
type SessionStats struct {
	DownloadedBytes int64 `json:"downloadedBytes"`
	UploadedBytes   int64 `json:"uploadedBytes"`
	FilesAdded      int   `json:"filesAdded"`
	SessionCount    int   `json:"sessionCount"`
	SecondsActive   int   `json:"secondsActive"`
}

// TransmissionSessionResponse represents session-get response
type TransmissionSessionResponse struct {
	Arguments SessionInfo `json:"arguments"`
	Result    string      `json:"result"`
}

// TransmissionStatsResponse represents session-stats response
type TransmissionStatsResponse struct {
	Arguments struct {
		CurrentStats    SessionStats `json:"current-stats"`
		CumulativeStats SessionStats `json:"cumulative-stats"`
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
