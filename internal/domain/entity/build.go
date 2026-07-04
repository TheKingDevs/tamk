package entity

type BuildPhase string

const (
	BuildPhaseAAPT2Compile  BuildPhase = "aapt2_compile"
	BuildPhaseAAPT2Link     BuildPhase = "aapt2_link"
	BuildPhaseKotlinCompile BuildPhase = "kotlin_compile"
	BuildPhaseD8            BuildPhase = "d8"
	BuildPhaseR8            BuildPhase = "r8"
	BuildPhasePackageDEX    BuildPhase = "package_dex"
	BuildPhaseZipalign      BuildPhase = "zipalign"
	BuildPhaseApkSign       BuildPhase = "apk_sign"
	BuildPhaseAABBuild      BuildPhase = "aab_build"
)

type BuildTarget string

const (
	BuildTargetAPK BuildTarget = "apk"
	BuildTargetAAB BuildTarget = "aab"
)

type BuildResult struct {
	Success  bool
	APKPath  string
	AABPath  string
	Phase    BuildPhase
	ErrorMsg string
}

type BuildCache struct {
	Hash     string
	PrevHash string
	Changed  bool
}
