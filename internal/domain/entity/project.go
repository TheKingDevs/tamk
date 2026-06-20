package entity

type ProjectType string

const (
	ProjectTypeWebApp  ProjectType = "webapp"
	ProjectTypeUIAPK   ProjectType = "ui_apk"
	ProjectTypeConsole ProjectType = "console"
)

type WebContentMode string

const (
	WebContentInternal WebContentMode = "internal"
	WebContentExternal WebContentMode = "external"
)

type Project struct {
	Name        string
	Type        ProjectType
	Version     string
	Author      string
	PackageName string

	WebURL       string
	WebMode      WebContentMode
	KeystorePath string

	MinSDK    int
	TargetSDK int

	CreatedAt string
	UpdatedAt string
}
