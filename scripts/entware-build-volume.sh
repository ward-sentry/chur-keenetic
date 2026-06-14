#!/bin/sh
set -eu

IMAGE="${ENTWARE_DOCKER_IMAGE:-chur-entware-build:20.04}"
DIST_DIR="${DIST_DIR:-dist/entware-ipk}"
TARGET_CONFIG="${ENTWARE_TARGET_CONFIG:-aarch64-3.10}"
VOLUME="${ENTWARE_DOCKER_VOLUME:-chur-entware-$TARGET_CONFIG}"
ENV_FILE="${ENV_FILE:-.env}"
CHUR_VERSION="${CHUR_VERSION:-$(cat VERSION)}"

if [ -f "$ENV_FILE" ]; then
	. "$ENV_FILE"
fi

AMNEZIAWG_TOOLS_COMMIT="${AMNEZIAWG_TOOLS_COMMIT:-5d6179a6d0842e98dfb349c28cf1bd8e4b9d1079}"
AMNEZIAWG_TOOLS_SHA256="${AMNEZIAWG_TOOLS_SHA256:-e79a3c7f2def315d052a3648b49058a268c4b63cdb5e082b696d2a4a0a2367f0}"
AMNEZIAWG_GO_SHORT_COMMIT=$(printf '%.7s' "${AMNEZIAWG_GO_COMMIT:-f4f4c999267437c3eb909e8d0e5278fb4596d9a7}")
CHUR_MANAGER_BINARY="${CHUR_MANAGER_BINARY:-/src/dist/chur_${CHUR_VERSION}_entware_${TARGET_CONFIG}}"
CHUR_AMNEZIAWG_GO_BINARY="${CHUR_AMNEZIAWG_GO_BINARY:-/src/dist/amneziawg-go_${AMNEZIAWG_GO_SHORT_COMMIT}_entware_${TARGET_CONFIG}}"

mkdir -p "$DIST_DIR"
find "$DIST_DIR" -maxdepth 1 -type f -name "*_${TARGET_CONFIG}.ipk" -delete
docker volume create "$VOLUME" >/dev/null

docker run --rm \
	-e HOST_UID="$(id -u)" \
	-e HOST_GID="$(id -g)" \
	-e HOME=/tmp \
	-e CHUR_VERSION="$CHUR_VERSION" \
	-e CHUR_MANAGER_BINARY="$CHUR_MANAGER_BINARY" \
	-e CHUR_AMNEZIAWG_GO_VERSION="$AMNEZIAWG_GO_SHORT_COMMIT" \
	-e CHUR_AMNEZIAWG_GO_BINARY="$CHUR_AMNEZIAWG_GO_BINARY" \
	-e CHUR_AMNEZIAWG_TOOLS_COMMIT="$AMNEZIAWG_TOOLS_COMMIT" \
	-e CHUR_AMNEZIAWG_TOOLS_SHA256="$AMNEZIAWG_TOOLS_SHA256" \
	-v "$VOLUME:/entware" \
	-v "$(pwd):/src" \
	-w /entware \
	"$IMAGE" \
	bash -lc '
		set -euo pipefail
		groupadd -g "$HOST_GID" churbuild 2>/dev/null || true
		useradd -u "$HOST_UID" -g "$HOST_GID" -M -d /tmp -s /bin/bash churbuild 2>/dev/null || true
		chown -R "$HOST_UID:$HOST_GID" /entware

		if [ ! -d /entware/.git ]; then
			su churbuild -s /bin/bash -c "git clone --depth 1 https://github.com/Entware/Entware /entware"
		fi

		mkdir -p /entware/package/chur
		rm -rf /entware/package/chur/chur
		rm -rf /entware/package/chur/chur-amneziawg
		rm -rf /entware/package/chur/chur-amneziawg-go
		rm -rf /entware/package/chur/chur-amneziawg-tools
		rm -rf /entware/package/chur/chur-keenetic
		cp -R /src/packaging/entware/chur-amneziawg /entware/package/chur/
		cp -R /src/packaging/entware/chur-amneziawg-go /entware/package/chur/
		cp -R /src/packaging/entware/chur-amneziawg-tools /entware/package/chur/
		cp -R /src/packaging/entware/chur-keenetic /entware/package/chur/
		chown -R "$HOST_UID:$HOST_GID" /entware/package/chur

		cat >/tmp/chur-entware-build.sh <<'"'"'EOS'"'"'
set -euo pipefail
cd /entware
test -x "$CHUR_MANAGER_BINARY"
test -x "$CHUR_AMNEZIAWG_GO_BINARY"
ENTWARE_DIR=/entware sh /src/scripts/entware-prepare-tree.sh

test -f .config || cp "configs/'"$TARGET_CONFIG"'.config" .config
./scripts/feeds update -a
./scripts/feeds install -a
sed -i '/CONFIG_PACKAGE_chur=/d' .config
sed -i '/CONFIG_PACKAGE_chur-amneziawg=/d' .config
sed -i '/CONFIG_PACKAGE_chur-amneziawg-go/d' .config
sed -i '/CONFIG_PACKAGE_chur-amneziawg-tools/d' .config
sed -i '/CONFIG_PACKAGE_chur-keenetic/d' .config
printf "%s\n" "CONFIG_PACKAGE_chur-amneziawg=m" >> .config
printf "%s\n" "CONFIG_PACKAGE_chur-amneziawg-go=m" >> .config
printf "%s\n" "CONFIG_PACKAGE_chur-amneziawg-tools=m" >> .config
printf "%s\n" "CONFIG_PACKAGE_chur-keenetic=m" >> .config
make defconfig
make tools/libdeflate/compile V=s
make tools/sed/compile V=s
make toolchain/kernel-headers/prepare V=s
ENTWARE_DIR=/entware sh /src/scripts/entware-prepare-tree.sh
make toolchain/install V=s
make package/chur/chur-amneziawg-go/compile V=s
make package/chur/chur-amneziawg-tools/compile V=s
make package/chur/chur-keenetic/compile V=s
make package/chur/chur-amneziawg/compile V=s
EOS
		chown "$HOST_UID:$HOST_GID" /tmp/chur-entware-build.sh
		su churbuild -s /bin/bash /tmp/chur-entware-build.sh

		mkdir -p /src/'"$DIST_DIR"'
		find /entware/bin -name "*chur-amneziawg-go*.ipk" -exec cp -f {} /src/'"$DIST_DIR"'/ \;
		find /entware/bin -name "*chur-amneziawg-tools*.ipk" -exec cp -f {} /src/'"$DIST_DIR"'/ \;
		find /entware/bin -name "*chur-amneziawg_*.ipk" -exec cp -f {} /src/'"$DIST_DIR"'/ \;
		find /entware/bin -name "*chur-keenetic*.ipk" -exec cp -f {} /src/'"$DIST_DIR"'/ \;
		chown -R "$HOST_UID:$HOST_GID" /src/'"$DIST_DIR"'
	'
