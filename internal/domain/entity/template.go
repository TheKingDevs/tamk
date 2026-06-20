package entity

type Template struct {
	Name         string
	SourcePath   string
	DestPath     string
	Placeholders map[string]string
}

type TemplateMapping struct {
	Dest     string
	Template string
	Internal bool
}
