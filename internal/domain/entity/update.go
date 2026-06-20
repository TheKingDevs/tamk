package entity

type UpdateLevel int

const (
	UpdateLevelOptional UpdateLevel = 0
	UpdateLevelPatch    UpdateLevel = 1
	UpdateLevelMinor    UpdateLevel = 2
	UpdateLevelMajor    UpdateLevel = 3
	UpdateLevelCritical UpdateLevel = 4
)

type UpdateInfo struct {
	CurrentVersion string
	LatestVersion  string
	Level          UpdateLevel
	ReleaseNotes   string
	DownloadURL    string
	ReleaseDate    string
}
