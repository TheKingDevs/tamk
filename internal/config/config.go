package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var (
	Version = "1.0.0"
	Commit  = "unknown"
	Date    = ""

	once     sync.Once
	instance *Config
)

// Get returns a singleton Config instance.
// Use this in CLI entry points where a fresh config is not needed.
func Get() *Config {
	once.Do(func() {
		instance = New()
	})
	return instance
}

type Environment string

const (
	EnvTermux  Environment = "termux"
	EnvDebian  Environment = "debian"
	EnvUbuntu  Environment = "ubuntu"
	EnvArch    Environment = "arch"
	EnvFedora  Environment = "fedora"
	EnvUnknown Environment = "unknown"
)

type Config struct {
	Version    string
	Env        Environment
	EnvType    string
	TAMKHome   string
	DevDir     string
	SDKPath    string
	Keystore   string
	PkgMgr     string
	ProjectDir string
}

func New() *Config {
	cfg := &Config{
		Version: Version,
		EnvType: os.Getenv("TAMK_ENV"),
	}
	if cfg.EnvType == "" {
		cfg.EnvType = "development"
	}
	cfg.detectEnvironment()
	cfg.resolvePaths()

	// Override keystore path from environment if set
	if ks := os.Getenv("TAMK_KEYSTORE"); ks != "" {
		cfg.Keystore = ks
	}

	return cfg
}

func ConfigFromEnv() *Config {
	cfg := New()
	if sdk := os.Getenv("TAMK_SDK_PATH"); sdk != "" {
		cfg.SDKPath = sdk
	}
	if devDir := os.Getenv("TAMK_DEV_DIR"); devDir != "" {
		cfg.DevDir = devDir
	}
	if projDir := os.Getenv("TAMK_PROJECT_DIR"); projDir != "" {
		cfg.ProjectDir = projDir
	}
	return cfg
}

func (c *Config) detectEnvironment() {
	if _, err := os.Stat("/data/data/com.termux"); err == nil {
		c.Env = EnvTermux
		c.PkgMgr = "pkg"
		return
	}

	f, err := os.Open("/etc/os-release")
	if err != nil {
		c.Env = EnvUnknown
		return
	}
	defer f.Close()

	scan := bufio.NewScanner(f)
	for scan.Scan() {
		line := scan.Text()
		switch {
		case strings.Contains(line, "ID=debian"):
			c.Env = EnvDebian
			c.PkgMgr = "apt"
		case strings.Contains(line, "ID=ubuntu"):
			c.Env = EnvUbuntu
			c.PkgMgr = "apt"
		case strings.Contains(line, "ID=arch") || strings.Contains(line, "ID=manjaro"):
			c.Env = EnvArch
			c.PkgMgr = "pacman"
		case strings.Contains(line, "ID=fedora"):
			c.Env = EnvFedora
			c.PkgMgr = "dnf"
		}
	}

	if c.Env == "" {
		c.Env = EnvUnknown
		c.PkgMgr = "apt"
	}
}

func (c *Config) resolvePaths() {
	if home := os.Getenv("TAMK_HOME"); home != "" {
		c.TAMKHome = home
	} else {
		execPath, err := os.Executable()
		if err == nil {
			c.TAMKHome = filepath.Dir(filepath.Dir(execPath))
		} else {
			cwd, _ := os.Getwd()
			c.TAMKHome = cwd
		}
	}

	c.DevDir = filepath.Join(c.TAMKHome, "development")
	c.SDKPath = filepath.Join(c.DevDir, "sdk", "android.jar")
	c.Keystore = filepath.Join(c.DevDir, "secret", "debug.keystore")
}

var templateDirCache sync.Map

func (c *Config) GetTemplateDir(projectType string) string {
	if cached, ok := templateDirCache.Load(projectType); ok {
		return cached.(string)
	}

	cwd, _ := os.Getwd()
	candidates := []string{
		filepath.Join(c.TAMKHome, "templates", projectType),
		filepath.Join(c.TAMKHome, "assets", "templates", projectType),
		filepath.Join(c.TAMKHome, "src", "templates", projectType),
		filepath.Join(cwd, "templates", projectType),
		filepath.Join(cwd, "assets", "templates", projectType),
	}

	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			templateDirCache.Store(projectType, dir)
			return dir
		}
	}

	if c.Env == EnvTermux {
		for _, p := range []string{"templates", filepath.Join("assets", "templates")} {
			termuxPath := filepath.Join("/data/data/com.termux/files/usr", "opt", "tamk", p, projectType)
			if info, err := os.Stat(termuxPath); err == nil && info.IsDir() {
				templateDirCache.Store(projectType, termuxPath)
				return termuxPath
			}
		}
	}

	result := filepath.Join(c.TAMKHome, "templates", projectType)
	templateDirCache.Store(projectType, result)
	return result
}

func (c *Config) GetProjectDir(name string) string {
	base := c.ProjectDir
	if base == "" {
		cwd, err := os.Getwd()
		if err == nil {
			base = cwd
		} else {
			base = c.TAMKHome
		}
	}
	return filepath.Join(base, name)
}

func (c *Config) IsTermux() bool {
	return c.Env == EnvTermux
}

func (c *Config) Validate() error {
	if c.TAMKHome == "" {
		return fmt.Errorf("TAMK_HOME not set")
	}
	if c.Version == "" {
		return fmt.Errorf("version not set")
	}
	if c.PkgMgr == "" {
		return fmt.Errorf("package manager not detected")
	}
	return nil
}

func (c *Config) DetectPackageName(author, name string) string {
	author = strings.ToLower(strings.TrimSpace(author))
	name = strings.ToLower(strings.TrimSpace(name))
	replacer := strings.NewReplacer("-", "", "_", "", " ", "")
	return "com." + replacer.Replace(author) + "." + replacer.Replace(name)
}

func SecurePath(base string, parts ...string) string {
	elem := append([]string{base}, parts...)
	p := filepath.Join(elem...)
	p = filepath.Clean(p)

	if !strings.HasPrefix(p, filepath.Clean(base)) {
		return filepath.Clean(base)
	}
	return p
}

func HomeDir() string {
	home, err := os.UserHomeDir()
	if err == nil {
		return home
	}
	if runtime.GOOS == "linux" {
		return "/root"
	}
	return "/tmp"
}
