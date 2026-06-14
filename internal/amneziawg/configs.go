package amneziawg

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	defaultConfigRoot = "/opt/etc/chur-keenetic"
	maxConfigBytes    = 256 * 1024
)

var (
	configNamePattern       = regexp.MustCompile(`^opkgtun[0-9]{1,2}$`)
	legacyConfigNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,14}$`)
	opkgTunIndexPattern     = regexp.MustCompile(`(?i)^opkgtun([0-9]{1,2})$`)
)

type Config struct {
	Name             string    `json:"name"`
	Description      string    `json:"description,omitempty"`
	Path             string    `json:"path"`
	Address          string    `json:"address,omitempty"`
	AddressCommented bool      `json:"addressCommented,omitempty"`
	DNS              string    `json:"dns,omitempty"`
	DNSCommented     bool      `json:"dnsCommented,omitempty"`
	MTU              string    `json:"mtu,omitempty"`
	Endpoint         string    `json:"endpoint,omitempty"`
	AllowedIP        string    `json:"allowedIp,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type CommandResult struct {
	Path      string `json:"path,omitempty"`
	Found     bool   `json:"found"`
	ExitCode  int    `json:"exitCode,omitempty"`
	Output    string `json:"output,omitempty"`
	Error     string `json:"error,omitempty"`
	TimedOut  bool   `json:"timedOut,omitempty"`
	Truncated bool   `json:"truncated,omitempty"`
}

type InterfaceStatus struct {
	Name         string        `json:"name"`
	ConfigPath   string        `json:"configPath"`
	ConfigExists bool          `json:"configExists"`
	Running      bool          `json:"running"`
	LinkFound    bool          `json:"linkFound"`
	Link         CommandResult `json:"link"`
	Show         CommandResult `json:"show"`
	CheckedAt    time.Time     `json:"checkedAt"`
}

type ActionResult struct {
	Name    string          `json:"name"`
	Action  string          `json:"action"`
	Command CommandResult   `json:"command,omitempty"`
	NDMC    []CommandResult `json:"ndmc,omitempty"`
	Status  InterfaceStatus `json:"status"`
}

type SaveConfigRequest struct {
	Name        string
	Description string
	Content     string
	MTU         string
}

type UpdateConfigRequest struct {
	Name        string
	Description string
	Content     string
	MTU         string
}

type UpdateConfigResult struct {
	Config    Config        `json:"config"`
	Restarted bool          `json:"restarted"`
	Stop      CommandResult `json:"stop,omitempty"`
	Start     CommandResult `json:"start,omitempty"`
}

type metadata struct {
	Description string `json:"description,omitempty"`
}

func ListConfigs(ctx context.Context) ([]Config, error) {
	dir := configsDir()
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return []Config{}, nil
	}
	if err != nil {
		return nil, err
	}

	configs := make([]Config, 0, len(entries))
	for _, entry := range entries {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".conf" {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".conf")
		content, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}

		config, err := parseConfig(name, string(content))
		if err != nil {
			config = Config{Name: name, Path: filepath.Join(dir, entry.Name())}
		}
		config.Description = loadMetadata(name).Description
		if info, err := entry.Info(); err == nil {
			config.UpdatedAt = info.ModTime().UTC()
			config.CreatedAt = config.UpdatedAt
		}
		configs = append(configs, config)
	}

	sort.Slice(configs, func(i, j int) bool {
		return configs[i].Name < configs[j].Name
	})
	return configs, nil
}

func SaveConfig(ctx context.Context, request SaveConfigRequest) (Config, error) {
	name, err := nextOrValidateName(ctx, request.Name)
	if err != nil {
		return Config{}, err
	}

	content := strings.TrimSpace(request.Content)
	if content == "" {
		return Config{}, errors.New("empty config")
	}
	if len(content) > maxConfigBytes {
		return Config{}, fmt.Errorf("config is too large: %d bytes", len(content))
	}
	content = normalizeConfig(content)
	if request.MTU != "" {
		content, err = setInterfaceMTU(content, request.MTU)
		if err != nil {
			return Config{}, err
		}
	}

	config, err := parseConfig(name, content)
	if err != nil {
		return Config{}, err
	}

	if ctx.Err() != nil {
		return Config{}, ctx.Err()
	}

	dir := configsDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Config{}, err
	}

	path := configPath(name)
	if err := os.WriteFile(path, []byte(content+"\n"), 0o600); err != nil {
		return Config{}, err
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return Config{}, err
	}
	description := cleanDescription(request.Description, name)
	if err := saveMetadata(name, metadata{Description: description}); err != nil {
		return Config{}, err
	}

	now := time.Now().UTC()
	config.Path = path
	config.Description = description
	config.CreatedAt = now
	config.UpdatedAt = now
	if info, err := os.Stat(path); err == nil {
		config.UpdatedAt = info.ModTime().UTC()
		config.CreatedAt = config.UpdatedAt
	}
	return config, nil
}

func UpdateConfig(ctx context.Context, request UpdateConfigRequest) (UpdateConfigResult, error) {
	name, err := validateExistingName(request.Name)
	if err != nil {
		return UpdateConfigResult{}, err
	}

	path := configPath(name)
	content := strings.TrimSpace(request.Content)
	if content == "" {
		existingContent, err := os.ReadFile(path)
		if err != nil {
			return UpdateConfigResult{}, err
		}
		content = string(existingContent)
	}
	if len(content) > maxConfigBytes {
		return UpdateConfigResult{}, fmt.Errorf("config is too large: %d bytes", len(content))
	}

	content = normalizeConfig(content)
	if request.MTU != "" {
		content, err = setInterfaceMTU(content, request.MTU)
		if err != nil {
			return UpdateConfigResult{}, err
		}
	}

	before, err := Status(ctx, name)
	if err != nil {
		return UpdateConfigResult{}, err
	}

	result := UpdateConfigResult{}
	if before.Running {
		result.Stop = run(ctx, 30*time.Second, "awg-quick", "down", path)
		if result.Stop.Error != "" || result.Stop.TimedOut {
			return result, fmt.Errorf("stop interface before update failed: %s", result.Stop.Error)
		}
	}

	config, err := SaveConfig(ctx, SaveConfigRequest{
		Name:        name,
		Description: request.Description,
		Content:     content,
	})
	if err != nil {
		return result, err
	}
	result.Config = config

	if before.Running {
		result.Restarted = true
		result.Start = run(ctx, 30*time.Second, "awg-quick", "up", path)
		if result.Start.Error != "" || result.Start.TimedOut {
			return result, fmt.Errorf("start interface after update failed: %s", result.Start.Error)
		}
	}

	return result, nil
}

func DeleteConfig(ctx context.Context, name string) (ActionResult, error) {
	name, err := validateExistingName(name)
	if err != nil {
		return ActionResult{}, err
	}

	status, err := Status(ctx, name)
	if err != nil {
		return ActionResult{}, err
	}

	result := ActionResult{
		Name:   name,
		Action: "delete",
		Status: status,
	}
	if status.Running {
		result.Command = run(ctx, 30*time.Second, "awg-quick", "down", configPath(name))
		if result.Command.Error != "" || result.Command.TimedOut {
			result.Status, _ = Status(ctx, name)
			return result, fmt.Errorf("stop interface before delete failed: %s", result.Command.Error)
		}
	}
	removeLinuxInterface(ctx, name)
	if lowerName := strings.ToLower(name); lowerName != name {
		removeLinuxInterface(ctx, lowerName)
	}
	result.NDMC = removeKeeneticInterface(ctx, name)

	if err := os.Remove(configPath(name)); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			result.Status, _ = Status(ctx, name)
			return result, nil
		}
		return result, err
	}
	_ = os.Remove(metadataPath(name))

	result.Status, _ = Status(ctx, name)
	return result, nil
}

func Start(ctx context.Context, name string) (ActionResult, error) {
	return changeState(ctx, name, "start", "up")
}

func Stop(ctx context.Context, name string) (ActionResult, error) {
	return changeState(ctx, name, "stop", "down")
}

func Status(ctx context.Context, name string) (InterfaceStatus, error) {
	name, err := validateExistingName(name)
	if err != nil {
		return InterfaceStatus{}, err
	}

	path := configPath(name)
	_, statErr := os.Stat(path)
	status := InterfaceStatus{
		Name:         name,
		ConfigPath:   path,
		ConfigExists: statErr == nil,
		CheckedAt:    time.Now().UTC(),
	}

	if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
		return status, statErr
	}
	if ctx.Err() != nil {
		return status, ctx.Err()
	}

	status.Link = run(ctx, 3*time.Second, "ip", "link", "show", "dev", name)
	status.LinkFound = status.Link.Found && status.Link.Error == "" && !status.Link.TimedOut
	status.Show = run(ctx, 3*time.Second, "awg", "show", name)
	status.Running = status.Show.Found && status.Show.Error == "" && !status.Show.TimedOut
	if !status.Running && status.LinkFound {
		status.Running = true
	}

	return status, nil
}

func ValidateName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if !configNamePattern.MatchString(name) {
		return "", fmt.Errorf("invalid interface name %q: use opkgtun0, opkgtun1, ... opkgtun99", name)
	}
	return name, nil
}

func validateExistingName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if !legacyConfigNamePattern.MatchString(name) {
		return "", fmt.Errorf("invalid interface name %q: use 1-15 letters, digits, _ or -", name)
	}
	return name, nil
}

func nextOrValidateName(ctx context.Context, name string) (string, error) {
	name = strings.TrimSpace(name)
	if name != "" {
		return ValidateName(name)
	}
	return NextName(ctx)
}

func NextName(ctx context.Context) (string, error) {
	used := map[int]bool{}
	entries, err := os.ReadDir(configsDir())
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	for _, entry := range entries {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".conf" {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".conf")
		if index, ok := opkgTunIndex(name); ok {
			used[index] = true
		}
	}
	for index := 0; index <= 99; index++ {
		if !used[index] {
			return fmt.Sprintf("opkgtun%d", index), nil
		}
	}
	return "", errors.New("no free opkgtun interface index")
}

func opkgTunIndex(name string) (int, bool) {
	matches := opkgTunIndexPattern.FindStringSubmatch(strings.TrimSpace(name))
	if len(matches) != 2 {
		return 0, false
	}
	var index int
	for _, ch := range matches[1] {
		index = index*10 + int(ch-'0')
	}
	return index, true
}

func parseConfig(name string, content string) (Config, error) {
	currentSection := ""
	sections := map[string]bool{}
	values := map[string]map[string]string{}
	commentedValues := map[string]map[string]string{}

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		isCommented := false
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			isCommented = true
			line = strings.TrimSpace(line[1:])
			if line == "" {
				continue
			}
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			if isCommented {
				continue
			}
			currentSection = strings.TrimSpace(line[1 : len(line)-1])
			if currentSection == "" {
				return Config{}, errors.New("empty section name")
			}
			sections[currentSection] = true
			if values[currentSection] == nil {
				values[currentSection] = map[string]string{}
			}
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok || currentSection == "" {
			continue
		}
		target := values
		if isCommented {
			target = commentedValues
		}
		if target[currentSection] == nil {
			target[currentSection] = map[string]string{}
		}
		target[currentSection][strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	if err := scanner.Err(); err != nil {
		return Config{}, err
	}

	if !sections["Interface"] {
		return Config{}, errors.New("missing [Interface] section")
	}
	if !sections["Peer"] {
		return Config{}, errors.New("missing [Peer] section")
	}
	if values["Interface"]["PrivateKey"] == "" {
		return Config{}, errors.New("missing Interface.PrivateKey")
	}
	if values["Peer"]["PublicKey"] == "" {
		return Config{}, errors.New("missing Peer.PublicKey")
	}
	if values["Peer"]["Endpoint"] == "" {
		return Config{}, errors.New("missing Peer.Endpoint")
	}

	path := configPath(name)
	address, addressCommented := configValue(values, commentedValues, "Interface", "Address")
	dns, dnsCommented := configValue(values, commentedValues, "Interface", "DNS")
	return Config{
		Name:             name,
		Description:      loadMetadata(name).Description,
		Path:             path,
		Address:          address,
		AddressCommented: addressCommented,
		DNS:              dns,
		DNSCommented:     dnsCommented,
		MTU:              values["Interface"]["MTU"],
		Endpoint:         values["Peer"]["Endpoint"],
		AllowedIP:        values["Peer"]["AllowedIPs"],
	}, nil
}

func normalizeConfig(content string) string {
	var lines []string
	inInterface := false
	interfaceTableSet := false

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			if inInterface && !interfaceTableSet {
				lines = append(lines, "Table = off")
				interfaceTableSet = true
			}
			inInterface = strings.EqualFold(strings.TrimSpace(trimmed[1:len(trimmed)-1]), "Interface")
		}
		if shouldDropEmptyAmneziaWGValue(trimmed) {
			continue
		}
		if normalizedDNS, ok := commentActiveDNSLine(line); ok {
			line = normalizedDNS
		}
		if inInterface && isTableLine(line) {
			interfaceTableSet = true
		}
		if normalizedAllowedIPs, ok := normalizeAllowedIPsLine(line); ok {
			line = normalizedAllowedIPs
		}
		lines = append(lines, line)
	}
	if inInterface && !interfaceTableSet {
		lines = append(lines, "Table = off")
	}

	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func setInterfaceMTU(content string, mtu string) (string, error) {
	mtu = strings.TrimSpace(mtu)
	if mtu == "" {
		return content, nil
	}
	if err := validateMTU(mtu); err != nil {
		return "", err
	}

	var lines []string
	inInterface := false
	interfaceFound := false
	mtuSet := false

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			if inInterface && !mtuSet {
				lines = append(lines, "MTU = "+mtu)
				mtuSet = true
			}
			inInterface = strings.EqualFold(strings.TrimSpace(trimmed[1:len(trimmed)-1]), "Interface")
			if inInterface {
				interfaceFound = true
			}
		}
		if inInterface && isMTULine(trimmed) {
			if !mtuSet {
				lines = append(lines, "MTU = "+mtu)
				mtuSet = true
			}
			continue
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	if !interfaceFound {
		return "", errors.New("missing [Interface] section")
	}
	if inInterface && !mtuSet {
		lines = append(lines, "MTU = "+mtu)
	}
	return strings.TrimSpace(strings.Join(lines, "\n")), nil
}

func validateMTU(mtu string) error {
	if len(mtu) < 3 || len(mtu) > 4 {
		return fmt.Errorf("invalid MTU %q: use 576-9000", mtu)
	}
	value := 0
	for _, ch := range mtu {
		if ch < '0' || ch > '9' {
			return fmt.Errorf("invalid MTU %q: use 576-9000", mtu)
		}
		value = value*10 + int(ch-'0')
	}
	if value < 576 || value > 9000 {
		return fmt.Errorf("invalid MTU %q: use 576-9000", mtu)
	}
	return nil
}

func isMTULine(line string) bool {
	if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
		return false
	}
	key, _, ok := strings.Cut(line, "=")
	return ok && strings.EqualFold(strings.TrimSpace(key), "MTU")
}

func isTableLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
		return false
	}

	key, _, ok := strings.Cut(trimmed, "=")
	return ok && strings.EqualFold(strings.TrimSpace(key), "Table")
}

func normalizeAllowedIPsLine(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
		return "", false
	}

	key, value, ok := strings.Cut(trimmed, "=")
	if !ok || !strings.EqualFold(strings.TrimSpace(key), "AllowedIPs") {
		return "", false
	}

	var ipv4Values []string
	for _, item := range strings.Split(value, ",") {
		item = strings.TrimSpace(item)
		if item == "" || strings.Contains(item, ":") {
			continue
		}
		ipv4Values = append(ipv4Values, item)
	}
	if len(ipv4Values) == 0 {
		return line, true
	}

	return "AllowedIPs = " + strings.Join(ipv4Values, ", "), true
}

func commentActiveDNSLine(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
		return "", false
	}

	key, value, ok := strings.Cut(trimmed, "=")
	if !ok || !strings.EqualFold(strings.TrimSpace(key), "DNS") {
		return "", false
	}
	return "# DNS = " + strings.TrimSpace(value), true
}

func shouldDropEmptyAmneziaWGValue(line string) bool {
	if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
		return false
	}

	key, value, ok := strings.Cut(line, "=")
	if !ok || strings.TrimSpace(value) != "" {
		return false
	}

	switch strings.TrimSpace(key) {
	case "I1", "I2", "I3", "I4", "I5":
		return true
	default:
		return false
	}
}

func changeState(ctx context.Context, name string, action string, awgQuickAction string) (ActionResult, error) {
	name, err := validateExistingName(name)
	if err != nil {
		return ActionResult{}, err
	}
	if action == "start" {
		if _, err := ValidateName(name); err != nil {
			return ActionResult{
				Name:   name,
				Action: action,
			}, fmt.Errorf("interface %q uses a legacy name; delete it and create a new one as opkgtun0, opkgtun1, ...", name)
		}
	}
	if _, err := os.Stat(configPath(name)); err != nil {
		return ActionResult{}, err
	}

	result := ActionResult{
		Name:   name,
		Action: action,
	}

	before, err := Status(ctx, name)
	if err != nil {
		return result, err
	}
	if action == "start" && before.Running {
		result.Status = before
		return result, nil
	}
	if action == "stop" && !before.Running {
		result.Status = before
		return result, nil
	}

	result.Command = run(ctx, 30*time.Second, "awg-quick", awgQuickAction, configPath(name))
	result.Status, _ = Status(ctx, name)
	if result.Command.Error != "" || result.Command.TimedOut {
		return result, fmt.Errorf("awg-quick %s failed: %s", awgQuickAction, result.Command.Error)
	}
	return result, nil
}

func removeKeeneticInterface(ctx context.Context, name string) []CommandResult {
	if _, err := exec.LookPath("ndmc"); err != nil {
		return nil
	}
	ndmsName := keeneticName(name)
	return []CommandResult{
		run(ctx, 10*time.Second, "ndmc", "-c", "no interface "+ndmsName),
		run(ctx, 10*time.Second, "ndmc", "-c", "system configuration save"),
	}
}

func removeLinuxInterface(ctx context.Context, name string) CommandResult {
	if name == "" {
		return CommandResult{}
	}
	if ipPath, err := exec.LookPath("ip"); err == nil {
		return run(ctx, 10*time.Second, ipPath, "link", "delete", name)
	}
	if _, err := os.Stat("/opt/sbin/ip"); err == nil {
		return run(ctx, 10*time.Second, "/opt/sbin/ip", "link", "delete", name)
	}
	return CommandResult{}
}

func run(parent context.Context, timeout time.Duration, name string, args ...string) CommandResult {
	path, lookPathErr := exec.LookPath(name)
	if lookPathErr != nil {
		return CommandResult{Found: false, Error: lookPathErr.Error()}
	}

	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, args...)
	output, err := cmd.CombinedOutput()

	result := CommandResult{
		Path:   path,
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

func configPath(name string) string {
	return filepath.Join(configsDir(), name+".conf")
}

func metadataPath(name string) string {
	return filepath.Join(configsDir(), name+".json")
}

func loadMetadata(name string) metadata {
	content, err := os.ReadFile(metadataPath(name))
	if err != nil {
		return metadata{}
	}
	var meta metadata
	if err := json.Unmarshal(content, &meta); err != nil {
		return metadata{}
	}
	return meta
}

func saveMetadata(name string, meta metadata) error {
	content, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(metadataPath(name), append(content, '\n'), 0o600)
}

func cleanDescription(description string, fallback string) string {
	description = strings.TrimSpace(description)
	if description == "" {
		description = fallback
	}
	description = strings.ReplaceAll(description, `"`, `'`)
	if len(description) > 48 {
		description = description[:48]
	}
	return description
}

func keeneticName(name string) string {
	index, ok := opkgTunIndex(name)
	if !ok {
		return name
	}
	return fmt.Sprintf("OpkgTun%d", index)
}

func configValue(values map[string]map[string]string, commentedValues map[string]map[string]string, section string, key string) (string, bool) {
	if value := values[section][key]; value != "" {
		return value, false
	}
	if value := commentedValues[section][key]; value != "" {
		return value, true
	}
	return "", false
}

func configsDir() string {
	root := os.Getenv("CHUR_CONFIG_DIR")
	if root == "" {
		root = defaultConfigRoot
	}
	return filepath.Join(root, "amneziawg")
}
