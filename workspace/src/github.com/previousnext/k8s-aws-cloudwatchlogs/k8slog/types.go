package k8slog

import "time"

// LogFile is used for K8s Log file utilities.
type LogFile struct {
	Namespace string
	Pod       string
	Container string
	path      string
}

// Log is for each line in the file.
type Log struct {
	Log    string    `json:"log"`
	Stream string    `json:"stream"`
	Time   time.Time `json:"time"`
}
