package model

// RootFolder - Stores struct of JSON response
type RootFolder []struct {
	Path       string `json:"path"`
	FreeSpace  int64  `json:"freeSpace"`
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

type Update []struct {
	Version     string `json:"version"`
	Branch      string `json:"branch"`
	ReleaseDate string `json:"releaseDate"`
	Installed   bool   `json:"installed"`
	Latest      bool   `json:"latest"`
	Hash        string `json:"hash"`
}

type DiskSpace []struct {
	Path       string `json:"path"`
	Label      string `json:"label"`
	FreeSpace  int64  `json:"freeSpace"`
	TotalSpace int64  `json:"totalSpace"`
}

type BlockList struct {
	TotalRecords int `json:"totalRecords"`
}

type Logs struct {
	TotalRecords int        `json:"totalRecords"`
	Records      []struct {
		Level   string `json:"level"`
		Message string `json:"message"`
		Time    string `json:"time"`
		Logger  string `json:"logger"`
	}`json:"records"`
}
