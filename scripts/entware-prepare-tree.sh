#!/bin/sh
set -eu

ENTWARE_DIR="${ENTWARE_DIR:-build/entware}"
GO_BOOTSTRAP_MAKEFILE="$ENTWARE_DIR/tools/go-bootstrap/Makefile"
GO_SRC_MAKEFILE="$ENTWARE_DIR/tools/go-src/Makefile"

if [ ! -f "$GO_BOOTSTRAP_MAKEFILE" ]; then
	echo "Missing Entware go-bootstrap recipe: $GO_BOOTSTRAP_MAKEFILE" >&2
	exit 1
fi

if [ ! -f "$GO_SRC_MAKEFILE" ]; then
	echo "Missing Entware go-src recipe: $GO_SRC_MAKEFILE" >&2
	exit 1
fi

# Entware commit 2dccac1 carries the amd64 Go 1.24.6 hash in a host-arch
# bootstrap recipe. On arm64 Linux builders the official Go archive has a
# different checksum.
OLD_HASH="bbca37cc395c974ffa4893ee35819ad23ebb27426df87af92e93a9ec66ef8712"
LINUX_ARM64_HASH="124ea6033a8bf98aa9fbab53e58d134905262d45a022af3a90b73320f3c3afd5"

case "$(uname -m)" in
	arm64|aarch64)
		if grep -q "PKG_VERSION:=1.24.6" "$GO_BOOTSTRAP_MAKEFILE" &&
			grep -q "PKG_HASH:=$OLD_HASH" "$GO_BOOTSTRAP_MAKEFILE"; then
			tmp="$GO_BOOTSTRAP_MAKEFILE.tmp"
			sed "s/PKG_HASH:=$OLD_HASH/PKG_HASH:=$LINUX_ARM64_HASH/" "$GO_BOOTSTRAP_MAKEFILE" > "$tmp"
			mv "$tmp" "$GO_BOOTSTRAP_MAKEFILE"
			echo "Patched Entware go-bootstrap Go 1.24.6 linux-arm64 hash"
		fi
		;;
	esac

if ! grep -q "GOTOOLCHAIN=local" "$GO_SRC_MAKEFILE"; then
	tmp="$GO_SRC_MAKEFILE.tmp"
	sed 's/GOTELEMETRY=off \\/GOTELEMETRY=off \\\n\t\t\tGOTOOLCHAIN=local \\/' "$GO_SRC_MAKEFILE" > "$tmp"
	mv "$tmp" "$GO_SRC_MAKEFILE"
	echo "Patched Entware go-src build to use GOTOOLCHAIN=local"
fi

for kbuild in "$ENTWARE_DIR"/build_dir/toolchain-*/linux-*/include/uapi/linux/*/Kbuild "$ENTWARE_DIR"/build_dir/toolchain-*/linux-*/include/uapi/linux/Kbuild; do
	[ -f "$kbuild" ] || continue
	dir=$(dirname "$kbuild")
	tmp="$kbuild.tmp"
	while IFS= read -r line; do
		case "$line" in
			"header-y += "*)
				header=${line#header-y += }
				if [ "${header%/}" = "$header" ] && [ ! -e "$dir/$header" ]; then
					lower=$(printf '%s' "$header" | tr '[:upper:]' '[:lower:]')
					if [ "$lower" != "$header" ] && [ -e "$dir/$lower" ]; then
						continue
					fi
				fi
				;;
		esac
		printf '%s\n' "$line"
	done < "$kbuild" > "$tmp"
	if ! cmp -s "$kbuild" "$tmp"; then
		mv "$tmp" "$kbuild"
		echo "Patched Linux 3.10 UAPI Kbuild case-collision entries in $kbuild"
	else
		rm -f "$tmp"
	fi
done
