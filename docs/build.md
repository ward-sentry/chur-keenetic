# Build Notes

Первый dev target:

```text
Keenetic Ultra NC-1812
KeenOS 5.0.12
Entware arch: aarch64-3.10 / aarch64-3.10_kn
```

## Build manager binaries

Версия Chur берется из файла `VERSION`. Upstream commits для runtime
components берутся из `.env`.

`chur-keenetic` пока не использует cgo, поэтому manager binary можно собрать обычным Go cross-compile:

```sh
make build_entware_aarch64-3.10
make build_entware_aarch64-3.10_kn
make build_entware_mips-3.4
make build_entware_mips-3.4_kn
make build_entware_mipsel-3.4
make build_entware_mipsel-3.4_kn
```

Собрать все manager targets:

```sh
make build_all
```

This is the fast build path. It builds only Chur manager binaries, does not run
Docker, and does not build `.ipk` packages.

Имена артефактов:

```text
dist/chur_{VERSION}_entware_aarch64-3.10
dist/chur_{VERSION}_entware_aarch64-3.10_kn
dist/chur_{VERSION}_entware_mips-3.4
dist/chur_{VERSION}_entware_mips-3.4_kn
dist/chur_{VERSION}_entware_mipsel-3.4
dist/chur_{VERSION}_entware_mipsel-3.4_kn
```

Smoke-test локально:

```sh
make test
GOCACHE="$PWD/.cache/go-build" GOMODCACHE="$PWD/.cache/go-mod" go run ./cmd/chur-keenetic -listen :8088
curl http://127.0.0.1:8088/api/health
curl http://127.0.0.1:8088/api/system
```

Do not use `go test ./...` from the repository root after submodules are initialized: it enters upstream `third_party/` trees. Use `make test`, which tests only Chur packages.

Run on Ultra:

```sh
scp -O dist/chur_$(cat VERSION)_entware_aarch64-3.10_kn root@192.168.88.1:/opt/bin/chur-keenetic
ssh root@192.168.88.1 -p 22
chmod +x /opt/bin/chur-keenetic
/opt/bin/chur-keenetic -listen 0.0.0.0:8088
```

`-O` важен для Keenetic/Entware: современный `scp` без этого флага пытается работать через SFTP и падает с ошибкой вида:

```text
sh: /opt/libexec/sftp-server: not found
scp: Connection closed
```

Then open:

```text
http://192.168.88.1:8088/
```

## Runtime Build Policy

`amneziawg-go` and `amneziawg-tools` must be built from upstream sources pinned by tag/commit. Existing `.ipk` files from the imported reference project are only documentation/reverse-engineering material.

Pinned upstream refs are documented in [upstream.md](upstream.md).

Entware buildroot flow for C runtime packages is documented in [entware-build.md](entware-build.md).

Build manager binaries plus the Entware `.ipk` for `chur-amneziawg-tools`:

```sh
make build_all_with_ipk
```

This is the heavy build path. It runs the Entware build inside Docker and
downloads the official `amneziawg-tools` upstream archive pinned by commit and
SHA256 from `.env`. A local `third_party/amneziawg-tools` checkout is
not required for this target.

It currently creates two packages:

```text
chur
chur-amneziawg
chur-amneziawg-go
chur-amneziawg-tools
chur-keenetic
```

`chur` installs only the manager and web UI. Runtime providers are optional.
For AmneziaWG the UI installs `chur-amneziawg`, which depends on
`chur-amneziawg-go` and `chur-amneziawg-tools`.

Install both local packages on Ultra:

```sh
scp -O dist/entware-ipk/chur-amneziawg-tools_1.0.20260223-1_aarch64-3.10.ipk root@192.168.88.1:/opt/tmp/
scp -O dist/entware-ipk/chur-keenetic_0.1.0-1_aarch64-3.10.ipk root@192.168.88.1:/opt/tmp/
ssh root@192.168.88.1 -p 22
opkg install /opt/tmp/chur-amneziawg-tools_1.0.20260223-1_aarch64-3.10.ipk /opt/tmp/chur-keenetic_0.1.0-1_aarch64-3.10.ipk
/opt/etc/init.d/S99chur-keenetic status
```

`chur-keenetic` installs:

```text
/opt/bin/chur-keenetic
/opt/etc/init.d/S99chur-keenetic
/opt/etc/chur-keenetic
/opt/var/log/chur-keenetic.log
```

The package post-install script starts the service automatically. Web UI:

```text
http://192.168.88.1:8088/
```

## Local opkg feed

Build a local feed from generated `.ipk` files:

```sh
make opkg_repo
make opkg_repo_serve
```

This serves:

```text
dist/opkg-repo/aarch64-3.10/Packages.gz
dist/opkg-repo/aarch64-3.10/*.ipk
```

On Ultra, add the local feed using the Mac IP address reachable from the router:

```sh
echo "src/gz chur http://MAC_IP:8090/aarch64-3.10" > /opt/etc/opkg/chur.conf
opkg update
opkg install chur
```

Then open the web UI and install AmneziaWG from the Runtime section when
needed:

```text
http://192.168.88.1:8088/
```

## Build amneziawg-go for Ultra

Build from pinned submodule commit from `.env`:

```sh
make build_runtime_amneziawg_go_entware_aarch64-3.10_kn
```

Artifact:

```text
dist/amneziawg-go_{UPSTREAM_SHORT_COMMIT}_entware_aarch64-3.10_kn
```

Install manually on Ultra for smoke-test:

```sh
scp -O dist/amneziawg-go_f4f4c99_entware_aarch64-3.10_kn root@192.168.88.1:/opt/bin/amneziawg-go
ssh root@192.168.88.1 -p 22
chmod +x /opt/bin/amneziawg-go
/opt/bin/amneziawg-go --help
```
