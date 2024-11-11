package model

// RootFolder - Stores struct of JSON response
type RootFolder []struct {
	Path      string `json:"path"`
	FreeSpace int64  `json:"freeSpace"`
	TotalSpace int64  `json:"totalSpace"`
}

// SystemStatus - Stores struct of JSON response
type SystemStatus struct {
	Version string `json:"version"`
	AppData string `json:"appData"`
	Branch  string `json:"branch"`
}

// Queue - Stores struct of JSON response
type Queue struct {
	Page          int            `json:"page"`
	PageSize      int            `json:"pageSize"`
	SortKey       string         `json:"sortKey"`
	SortDirection string         `json:"sortDirection"`
	TotalRecords  int            `json:"totalRecords"`
	Records       []QueueRecords `json:"records"`
}

// QueueRecords - Stores struct of JSON response
type QueueRecords struct {
	Size                  float64 `json:"size"`
	Title                 string  `json:"title"`
	Status                string  `json:"status"`
	TrackedDownloadStatus string  `json:"trackedDownloadStatus"`
	TrackedDownloadState  string  `json:"trackedDownloadState"`
	StatusMessages        []struct {
		Title    string   `json:"title"`
		Messages []string `json:"messages"`
	} `json:"statusMessages"`
	ErrorMessage string `json:"errorMessage"`
}

// History - Stores struct of JSON response
type History struct {
	TotalRecords int `json:"totalRecords"`
}

type SystemHealth []SystemHealthMessage

// SystemHealth - Stores struct of JSON response
type SystemHealthMessage struct {
	Source  string `json:"source"`
	Type    string `json:"type"`
	Message string `json:"message"`
	WikiURL string `json:"wikiUrl"`
}

type DownloadClient []struct {
	Name                     string `json:"name"`
	Protocol                 string `json:"protocol"`
	Enable                   bool   `json:"enable"`
	Priority                 int    `json:"priority"`
	RemoveCompletedDownloads bool   `json:"removeCompletedDownloads"`
	RemoveFailedDownloads    bool   `json:"removeFailedDownloads"`
	Implementation           string `json:"implementation"`
}

type ArrIndexer []struct {
	Name                    string `json:"name"`
	EnableRss               bool   `json:"enableRss"`
	EnableAutomaticSearch   bool   `json:"enableAutomaticSearch"`
	EnableInteractiveSearch bool   `json:"enableInteractiveSearch"`
	SupportsRss             bool   `json:"supportsRss"`
	SupportsSearch          bool   `json:"supportsSearch"`
	Protocol                string `json:"protocol"`
	Implementation          string `json:"implementation"`
	Priority                int    `json:"priority"`

	// Fields   []struct {
	// 	Name string `json:"name"`
	// 	// Value has multiple types, depending on the field, so it
	// 	// must be typecast at the call site.
	// 	Value interface{} `json:"value"`
	// } `json:"fields"`
}

/*
 "version": "2.0.0.548",
	"branch": "nightly",
	"releaseDate": "2024-04-09T03:02:26Z",
	"fileName": "Whisparr.develop.2.0.0.548.linux-musl-core-x64.tar.gz",
	"url": "https://dev.azure.com/Servarr/Whisparr/_apis/build/builds/3206/artifacts?artifactName=Packages&fileId=9A9838C750E18D5725BF557A516789D490BB1A12BFCB46DCB589BB61174D0BC002&fileName=Whisparr.develop.2.0.0.548.linux-musl-core-x64.tar.gz&api-version=5.1",
	"installed": true,
	"installedOn": "2024-05-27T21:35:29Z",
	"installable": false,
	"latest": true,
	"changes": {
			"new": [],
			"fixed": [
					"New TPDB domain"
			]
	},
	"hash": "1073e3854e5eaad09890a5b060c8037700b0c24d46054e61818ea7e5348a7b52"
*/

type Update []struct {
	Version      string `json:"version"`
	Branch       string `json:"branch"`
	ReleaseDate  string `json:"releaseDate"`
	Installed    bool   `json:"installed"`
	Latest       bool   `json:"latest"`
	Hash         string `json:"hash"`
}

type DiskSpace []struct {
	Path      string `json:"path"`
	Label     string `json:"label"`
	FreeSpace int64  `json:"freeSpace"`
	TotalSpace int64  `json:"totalSpace"`
}

type BlockList struct {
	TotalRecords int `json:"totalRecords"`
}
