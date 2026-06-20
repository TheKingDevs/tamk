package errors

import "errors"

var (
	ErrInvalidURL            = errors.New("invalid URL (must be http/https)")
	ErrProjectNotFound       = errors.New("project not found")
	ErrSDKNotFound           = errors.New("SDK not configured (run tamk --setup)")
	ErrKeystoreNotFound      = errors.New("keystore not found")
	ErrKeystoreInvalidPass   = errors.New("keystore password incorrect")
	ErrBuildFailed           = errors.New("build failed")
	ErrTemplateNotFound      = errors.New("template not found")
	ErrPathTraversal         = errors.New("path traversal detected")
	ErrNotAWebAppProject     = errors.New("not a WebApp project")
	ErrAPKBaseNotFound       = errors.New("APK base not found (build the project first)")
	ErrUpdateCheckFailed     = errors.New("update check failed")
	ErrDeviceNotConnected    = errors.New("no device connected")
	ErrWatchdogNotAvailable  = errors.New("file watcher not available")
	ErrWebSocketNotAvailable = errors.New("websocket not available")
)

type BuildError struct {
	Phase string
	Err   error
}

func (e *BuildError) Error() string {
	return "build failed at phase " + e.Phase + ": " + e.Err.Error()
}

func (e *BuildError) Unwrap() error {
	return e.Err
}
