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

type UIFramework string

const (
	UIFrameworkXML     UIFramework = "xml"
	UIFrameworkCompose UIFramework = "compose"
)

type Project struct {
	Name        string
	Type        ProjectType
	Version     string
	Author      string
	PackageName string

	WebURL       string
	WebMode      WebContentMode
	UIFramework  UIFramework
	KeystorePath string

	MinSDK    int
	TargetSDK int

	Security SecurityConfig `json:"security,omitempty"`

	CreatedAt string
	UpdatedAt string
}
