package config

import "fmt"

// DoctorIssue describes a single problem found in a profile by Doctor.
type DoctorIssue struct {
	// Profile is the name of the affected profile.
	Profile string
	// Jump is empty for the main profile identity_file, or
	// "jump[N]" for an Nth jump host identity_file.
	Jump string
	// Err is the underlying problem (missing file, insecure perms, etc.).
	Err error
}

// Doctor walks every profile in cfg and reports problems with identity
// files. It covers identity files that Profile.Validate also checks, plus
// those reachable only via the existing config (e.g. hand-edited YAML,
// imported SSH configs) where Validate was never invoked.
//
// Doctor never returns an error itself; it always returns the full list
// of issues so callers can show everything in one pass.
func Doctor(cfg *Config) []DoctorIssue {
	var issues []DoctorIssue
	for _, p := range cfg.Profiles {
		if p.IdentityFile != "" {
			if err := ValidateIdentityFile(ExpandTilde(p.IdentityFile)); err != nil {
				issues = append(issues, DoctorIssue{Profile: p.Name, Err: err})
			}
		}
		for i, j := range p.JumpHosts {
			if j.IdentityFile == "" {
				continue
			}
			if err := ValidateIdentityFile(ExpandTilde(j.IdentityFile)); err != nil {
				issues = append(issues, DoctorIssue{
					Profile: p.Name,
					Jump:    fmt.Sprintf("jump[%d]", i),
					Err:     err,
				})
			}
		}
	}
	return issues
}
