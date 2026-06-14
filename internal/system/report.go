package system

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/ward-sentry/chur-keenetic/internal/buildinfo"
)

type Report struct {
	OS          string         `json:"os"`
	Arch        string         `json:"arch"`
	GoVersion   string         `json:"goVersion"`
	Build       Build          `json:"build"`
	Hostname    string         `json:"hostname"`
	Paths       Paths          `json:"paths"`
	Commands    map[string]Cmd `json:"commands"`
	Keenetic    Keenetic       `json:"keenetic"`
	Entware     Entware        `json:"entware"`
	Runtime     Runtime        `json:"runtime"`
	CollectedAt time.Time      `json:"collectedAt"`
}

type Paths struct {
	OptExists bool `json:"optExists"`
}

type Build struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

type Cmd struct {
	Path      string `json:"path,omitempty"`
	Found     bool   `json:"found"`
	ExitCode  int    `json:"exitCode,omitempty"`
	Output    string `json:"output,omitempty"`
	Error     string `json:"error,omitempty"`
	TimedOut  bool   `json:"timedOut,omitempty"`
	Truncated bool   `json:"truncated,omitempty"`
}

type Keenetic struct {
	NDMCFound bool `json:"ndmcFound"`
	Version   Cmd  `json:"version"`
}

type Entware struct {
	OpkgFound    bool `json:"opkgFound"`
	Architecture Cmd  `json:"architecture"`
}

type Runtime struct {
	AmneziaWG AmneziaWGRuntime `json:"amneziawg"`
}

type AmneziaWGRuntime struct {
	Ready     bool `json:"ready"`
	Go        Cmd  `json:"go"`
	Tools     Cmd  `json:"tools"`
	Quick     Cmd  `json:"quick"`
	Missing   int  `json:"missing"`
	Installed int  `json:"installed"`
	Required  int  `json:"required"`
}

type RuntimeInstallResult struct {
	Provider string           `json:"provider"`
	Update   Cmd              `json:"update"`
	Install  Cmd              `json:"install"`
	Runtime  AmneziaWGRuntime `json:"runtime"`
	Ready    bool             `json:"ready"`
}

func Collect(ctx context.Context) Report {
	hostname, _ := os.Hostname()
	_, optErr := os.Stat("/opt")

	commands := map[string]Cmd{
		"ndmc": commandPath("ndmc"),
		"opkg": commandPath("opkg"),
		"ip":   commandPath("ip"),
		"curl": commandPath("curl"),
	}

	report := Report{
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		GoVersion: runtime.Version(),
		Build: Build{
			Version: buildinfo.Version,
			Commit:  buildinfo.Commit,
		},
		Hostname: hostname,
		Paths: Paths{
			OptExists: optErr == nil,
		},
		Commands:    commands,
		CollectedAt: time.Now().UTC(),
	}

	report.Keenetic.NDMCFound = commands["ndmc"].Found
	if report.Keenetic.NDMCFound {
		report.Keenetic.Version = run(ctx, 3*time.Second, "ndmc", "-c", "show version")
	}

	report.Entware.OpkgFound = commands["opkg"].Found
	if report.Entware.OpkgFound {
		report.Entware.Architecture = run(ctx, 3*time.Second, "opkg", "print-architecture")
	}

	report.Runtime.AmneziaWG = collectAmneziaWGRuntime()

	return report
}

func collectAmneziaWGRuntime() AmneziaWGRuntime {
	runtime := AmneziaWGRuntime{
		Go:       commandPath("amneziawg-go"),
		Tools:    commandPath("awg"),
		Quick:    commandPath("awg-quick"),
		Required: 3,
	}

	for _, cmd := range []Cmd{runtime.Go, runtime.Tools, runtime.Quick} {
		if cmd.Found {
			runtime.Installed++
		}
	}

	runtime.Missing = runtime.Required - runtime.Installed
	runtime.Ready = runtime.Missing == 0

	return runtime
}

func InstallAmneziaWG(ctx context.Context) RuntimeInstallResult {
	result := RuntimeInstallResult{
		Provider: "amneziawg",
	}

	if !commandPath("opkg").Found {
		result.Install = Cmd{
			Found: false,
			Error: "opkg not found",
		}
		result.Runtime = collectAmneziaWGRuntime()
		result.Ready = result.Runtime.Ready
		return result
	}

	result.Update = run(ctx, 60*time.Second, "opkg", "update")
	if result.Update.Error == "" && !result.Update.TimedOut {
		result.Install = run(ctx, 120*time.Second, "opkg", "install", "chur-amneziawg")
	}

	result.Runtime = collectAmneziaWGRuntime()
	result.Ready = result.Runtime.Ready
	return result
}

func commandPath(name string) Cmd {
	path, err := exec.LookPath(name)
	if err != nil {
		return Cmd{Found: false, Error: err.Error()}
	}
	return Cmd{Found: true, Path: path}
}

func run(parent context.Context, timeout time.Duration, name string, args ...string) Cmd {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()

	result := Cmd{
		Found:  true,
		Output: trimOutput(string(output), 4096),
	}
	result.Truncated = len(output) > 4096

	if ctx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
	}

	if err != nil {
		result.Error = err.Error()
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
	}

	return result
}

func trimOutput(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) <= max {
		return value
	}
	return value[:max]
}
