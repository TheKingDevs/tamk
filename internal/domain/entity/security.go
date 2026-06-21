package entity

// SecurityLevel defines the intensity of security protection.
type SecurityLevel string

const (
	SecurityLevelNone     SecurityLevel = "none"
	SecurityLevelBasic    SecurityLevel = "basic"
	SecurityLevelStandard SecurityLevel = "standard"
	SecurityLevelMaximum  SecurityLevel = "maximum"
)

// SecurityConfig holds all security settings for a project.
type SecurityConfig struct {
	Enabled  bool          `json:"enabled"`
	Level    SecurityLevel `json:"level"`
	Guardian GuardianConfig `json:"guardian"`
}

// GuardianConfig defines which security checks to enable.
type GuardianConfig struct {
	AntiDebug    bool `json:"anti_debug"`
	AntiRoot     bool `json:"anti_root"`
	AntiFrida    bool `json:"anti_frida"`
	AntiEmulator bool `json:"anti_emulator"`
	Integrity    bool `json:"integrity"`
	RASP         bool `json:"rasp"`
}

// DefaultSecurityConfig returns a standard security configuration.
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		Enabled: false,
		Level:   SecurityLevelNone,
		Guardian: GuardianConfig{
			AntiDebug:    false,
			AntiRoot:     false,
			AntiFrida:    false,
			AntiEmulator: false,
			Integrity:    false,
			RASP:         false,
		},
	}
}

// SecurityConfigForLevel returns a configuration preset for the given level.
func SecurityConfigForLevel(level SecurityLevel) SecurityConfig {
	switch level {
	case SecurityLevelBasic:
		return SecurityConfig{
			Enabled: true,
			Level:   level,
			Guardian: GuardianConfig{
				AntiDebug:    true,
				AntiRoot:     false,
				AntiFrida:    false,
				AntiEmulator: false,
				Integrity:    false,
				RASP:         false,
			},
		}
	case SecurityLevelStandard:
		return SecurityConfig{
			Enabled: true,
			Level:   level,
			Guardian: GuardianConfig{
				AntiDebug:    true,
				AntiRoot:     true,
				AntiFrida:    true,
				AntiEmulator: false,
				Integrity:    true,
				RASP:         false,
			},
		}
	case SecurityLevelMaximum:
		return SecurityConfig{
			Enabled: true,
			Level:   level,
			Guardian: GuardianConfig{
				AntiDebug:    true,
				AntiRoot:     true,
				AntiFrida:    true,
				AntiEmulator: true,
				Integrity:    true,
				RASP:         true,
			},
		}
	default:
		return DefaultSecurityConfig()
	}
}
