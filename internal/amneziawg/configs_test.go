package amneziawg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validConfig = `[Interface]
PrivateKey = private
Jc = 6

[Peer]
PublicKey = public
Endpoint = 127.0.0.1:12345
AllowedIPs = 0.0.0.0/0
`

func TestParseConfigReadsCommentedAddressAndDNS(t *testing.T) {
	t.Setenv("CHUR_CONFIG_DIR", t.TempDir())

	config, err := parseConfig("office", `[Interface]
# Address = 10.8.1.14/32
; DNS = 1.1.1.1, 1.0.0.1
PrivateKey = private

[Peer]
PublicKey = public
Endpoint = 127.0.0.1:12345
AllowedIPs = 0.0.0.0/0
`)
	if err != nil {
		t.Fatalf("parseConfig returned error: %v", err)
	}
	if config.Address != "10.8.1.14/32" || !config.AddressCommented {
		t.Fatalf("unexpected address: %q commented=%v", config.Address, config.AddressCommented)
	}
	if config.DNS != "1.1.1.1, 1.0.0.1" || !config.DNSCommented {
		t.Fatalf("unexpected dns: %q commented=%v", config.DNS, config.DNSCommented)
	}
}

func TestParseConfigPrefersActiveAddressAndDNS(t *testing.T) {
	t.Setenv("CHUR_CONFIG_DIR", t.TempDir())

	config, err := parseConfig("office", `[Interface]
# Address = 10.8.1.14/32
Address = 10.8.1.15/32
; DNS = 1.1.1.1
DNS = 8.8.8.8
PrivateKey = private

[Peer]
PublicKey = public
Endpoint = 127.0.0.1:12345
AllowedIPs = 0.0.0.0/0
`)
	if err != nil {
		t.Fatalf("parseConfig returned error: %v", err)
	}
	if config.Address != "10.8.1.15/32" || config.AddressCommented {
		t.Fatalf("unexpected address: %q commented=%v", config.Address, config.AddressCommented)
	}
	if config.DNS != "8.8.8.8" || config.DNSCommented {
		t.Fatalf("unexpected dns: %q commented=%v", config.DNS, config.DNSCommented)
	}
}

func TestParseConfigUsesInterfaceNameForPath(t *testing.T) {
	t.Setenv("CHUR_CONFIG_DIR", t.TempDir())

	config, err := parseConfig("home_vpn", validConfig)
	if err != nil {
		t.Fatalf("parseConfig returned error: %v", err)
	}
	if config.Name != "home_vpn" {
		t.Fatalf("unexpected name: %q", config.Name)
	}
	if got, want := config.Path[len(config.Path)-len("home_vpn.conf"):], "home_vpn.conf"; got != want {
		t.Fatalf("unexpected path suffix: got %q want %q", got, want)
	}
}

func TestValidateNameAllowsLinuxInterfaceNames(t *testing.T) {
	for _, name := range []string{"opkgtun0", "opkgtun1", "opkgtun99"} {
		if _, err := ValidateName(name); err != nil {
			t.Fatalf("ValidateName(%q) returned error: %v", name, err)
		}
	}
}

func TestValidateNameRejectsUnsafeNames(t *testing.T) {
	for _, name := range []string{"", "srv0", "ServerOne", "OpkgTun0", "OpkgTun", "opkgtun100", "with.dot", "with space", "home_vpn", "home-vpn", "../awg0", "1Server"} {
		if _, err := ValidateName(name); err == nil {
			t.Fatalf("ValidateName(%q) returned nil error", name)
		}
	}
}

func TestValidateExistingNameAllowsLegacyConfigNames(t *testing.T) {
	for _, name := range []string{"server_one", "home-vpn", "ServerOne", "OpkgTun0"} {
		if _, err := validateExistingName(name); err != nil {
			t.Fatalf("validateExistingName(%q) returned error: %v", name, err)
		}
	}
}

func TestNextNameUsesFirstFreeOpkgTunIndex(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CHUR_CONFIG_DIR", root)
	dir := filepath.Join(root, "amneziawg")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	for _, filename := range []string{"opkgtun0.conf", "OpkgTun1.conf", "server_one.conf"} {
		if err := os.WriteFile(filepath.Join(dir, filename), []byte(validConfig), 0o600); err != nil {
			t.Fatalf("WriteFile(%q) returned error: %v", filename, err)
		}
	}

	name, err := NextName(t.Context())
	if err != nil {
		t.Fatalf("NextName returned error: %v", err)
	}
	if name != "opkgtun2" {
		t.Fatalf("unexpected name: got %q want opkgtun2", name)
	}
}

func TestNormalizeConfigDropsEmptyAmneziaWGValues(t *testing.T) {
	content := normalizeConfig(`[Interface]
PrivateKey = private
I1 = <payload>
I2 =
I3 =   
DNS = 1.1.1.1, 1.0.0.1
Jc = 6

[Peer]
PublicKey = public
Endpoint = 127.0.0.1:12345
AllowedIPs = 0.0.0.0/0, ::/0
`)

	if strings.Contains(content, "I2") || strings.Contains(content, "I3") {
		t.Fatalf("empty I values were not dropped:\n%s", content)
	}
	if !strings.Contains(content, "I1 = <payload>") {
		t.Fatalf("non-empty I1 was dropped:\n%s", content)
	}
	if !strings.Contains(content, "# DNS = 1.1.1.1, 1.0.0.1") {
		t.Fatalf("DNS was not commented:\n%s", content)
	}
	if !strings.Contains(content, "AllowedIPs = 0.0.0.0/0") || strings.Contains(content, "::/0") {
		t.Fatalf("AllowedIPs was not normalized to IPv4-only:\n%s", content)
	}
	if !strings.Contains(content, "Table = off") {
		t.Fatalf("Table = off was not added:\n%s", content)
	}
}

func TestSaveConfigPreservesDNSForUIButCommentsItForAwgQuick(t *testing.T) {
	t.Setenv("CHUR_CONFIG_DIR", t.TempDir())

	config, err := SaveConfig(t.Context(), SaveConfigRequest{
		Name:        "opkgtun0",
		Description: "Office",
		Content: `[Interface]
Address = 10.8.1.14/32
DNS = 1.1.1.1, 1.0.0.1
PrivateKey = private

[Peer]
PublicKey = public
Endpoint = 127.0.0.1:12345
AllowedIPs = 0.0.0.0/0
`,
	})
	if err != nil {
		t.Fatalf("SaveConfig returned error: %v", err)
	}
	if config.DNS != "1.1.1.1, 1.0.0.1" || !config.DNSCommented {
		t.Fatalf("unexpected DNS: %q commented=%v", config.DNS, config.DNSCommented)
	}
	if config.Description != "Office" {
		t.Fatalf("unexpected description: %q", config.Description)
	}
}

func TestNormalizeConfigKeepsExistingTableSetting(t *testing.T) {
	content := normalizeConfig(`[Interface]
PrivateKey = private
Table = 123

[Peer]
PublicKey = public
Endpoint = 127.0.0.1:12345
AllowedIPs = 0.0.0.0/0
`)

	if strings.Contains(content, "Table = off") {
		t.Fatalf("unexpected Table = off:\n%s", content)
	}
	if !strings.Contains(content, "Table = 123") {
		t.Fatalf("existing Table setting was not preserved:\n%s", content)
	}
}

func TestSetInterfaceMTUReplacesExistingValue(t *testing.T) {
	content, err := setInterfaceMTU(`[Interface]
PrivateKey = private
MTU = 1420
Table = off

[Peer]
PublicKey = public
Endpoint = 127.0.0.1:12345
AllowedIPs = 0.0.0.0/0
`, "1280")
	if err != nil {
		t.Fatalf("setInterfaceMTU returned error: %v", err)
	}
	if !strings.Contains(content, "MTU = 1280") {
		t.Fatalf("new MTU was not written:\n%s", content)
	}
	if strings.Contains(content, "MTU = 1420") {
		t.Fatalf("old MTU was not replaced:\n%s", content)
	}
}

func TestSetInterfaceMTUAddsMissingValue(t *testing.T) {
	content, err := setInterfaceMTU(`[Interface]
PrivateKey = private
Table = off

[Peer]
PublicKey = public
Endpoint = 127.0.0.1:12345
AllowedIPs = 0.0.0.0/0
`, "1376")
	if err != nil {
		t.Fatalf("setInterfaceMTU returned error: %v", err)
	}
	if !strings.Contains(content, "MTU = 1376\n[Peer]") {
		t.Fatalf("MTU was not added before [Peer]:\n%s", content)
	}
}
