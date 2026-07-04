package config

// Project files
const (
	ConfigFileName     = "tamk.config"
	BuildCacheFileName = ".build_cache"
	ManifestFileName   = "AndroidManifest.xml"
	DexFileName        = "classes.dex"
)

// Directory names
const (
	SecretDirName      = "secret"
	DevelopmentDirName = "development"
	SrcDirName         = "src"
	MainDirName        = "main"
	AssetsDirName      = "assets"
	KotlinDirName      = "kotlin"
	ResDirName         = "res"
	ObjDirName         = "obj"
	GenDirName         = "gen"
	CacheDirName       = "cache"
	ExtractDirName     = "extract"
)

// Keystore files
const (
	ProjectKeystoreName = "project.keystore"
	DebugKeystoreName   = "debug.keystore"
)

// APK artifacts
const (
	ApkName         = "app.apk"
	UnsignedApkName = "app-unsigned.apk"
)

// Default values
const (
	DefaultMinSdkVersion = "21"
	DefaultMinSDK        = 21
	DefaultTargetSDK     = 30
	ReleaseEnv           = "release"
	DefaultEnvType       = "development"
	DefaultDevPort       = "8765"
)

// Keystore
const (
	KeystorePassPrefix = "pass:"
	KeyAlgorithm       = "RSA"
	KeySize            = "2048"
	KeyValidity        = "10000"
)

// Build artifacts
const (
	ResZipName    = "res.zip"
	ProtoAPKName  = "base.apk"
	AABName       = "bundle.aab"
	ModuleZipName = "module.zip"
	AABModuleDir  = "assets/cache/aab-module"
)
