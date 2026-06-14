# Entware Build

This document describes the clean path for building C runtime components with the official Entware build system.

## Why

`amneziawg-tools` is a C project. Building it with a random local compiler can produce binaries that do not match Entware ABI or paths. The clean path is:

```text
official Entware buildroot
  -> Entware target config
  -> Chur package recipe
  -> .ipk
```

The reference Entware package `package/network/utils/wireguard-tools` is used as the model.

## Docker Volume Entware Tree

The Entware tree is intentionally not committed. The local macOS path uses a
Linux-native Docker volume named `chur-entware-aarch64-3.10`.

```sh
make entware_docker_image
make entware_build_ipk_volume
```

The first target is:

```text
aarch64-3.10
```

Keenetic-specific package arch on Ultra is:

```text
aarch64-3.10_kn
```

The base Entware tree contains `configs/aarch64-3.10.config`.

## Build on Linux Host

The Entware/OpenWrt-style build system should be run on a Linux build host or container.

### Docker / OrbStack

Build the Linux build image:

```sh
make entware_docker_image
```

The build image currently uses Ubuntu 20.04 because Entware prereq checks still require Python 2.7.

Build Chur Entware packages:

```sh
make entware_build_ipk_volume
```

This builds Entware under a Linux-native Docker volume named
`chur-entware-aarch64-3.10` and copies resulting `.ipk` files into:

```text
dist/entware-ipk
```

On arm64 build hosts, the volume build first runs:

```sh
sh scripts/entware-prepare-tree.sh
```

At Entware commit `2dccac1`, `tools/go-bootstrap` uses Go `1.24.6`
and contains the `linux-amd64` checksum in a host-architecture recipe.
For `linux-arm64`, the official `go.dev` checksum for
`go1.24.6.linux-arm64.tar.gz` is:

```text
124ea6033a8bf98aa9fbab53e58d134905262d45a022af3a90b73320f3c3afd5
```

The prepare script applies only this local checkout fix before the Entware
build runs.

The same prepare step also sets `GOTOOLCHAIN=local` for Entware `go-src`.
Without it, bootstrap Go `1.24.6` may try to auto-download an intermediate
toolchain while building Go `1.26.1`, which breaks hermetic container builds.

On macOS/OrbStack bind mounts, Linux 3.10 kernel headers can also hit
case-collision exports such as `xt_CONNMARK.h` versus `xt_connmark.h`.
The build target runs `toolchain/kernel-headers/prepare`, applies the local
Kbuild cleanup, and then continues with `toolchain/install`.

The full `glibc` build also creates case-sensitive pairs such as `stamp.os`
and `stamp.oS`. That cannot be represented correctly on a case-insensitive
macOS bind mount, so the volume target is the clean local path.

## Package Behavior

`chur-amneziawg-tools` is built from the official upstream archive pinned in
`.env`:

```text
https://github.com/amnezia-vpn/amneziawg-tools/archive/5d6179a6d0842e98dfb349c28cf1bd8e4b9d1079.tar.gz
```

The package recipe receives both commit and SHA256 from `.env`:

```text
commit: 5d6179a6d0842e98dfb349c28cf1bd8e4b9d1079
sha256: e79a3c7f2def315d052a3648b49058a268c4b63cdb5e082b696d2a4a0a2367f0
```

It installs:

```text
/opt/bin/awg
/opt/bin/awg-quick
/opt/etc/amnezia/amneziawg
/opt/var/run/amneziawg
```

Adjustments applied during `Build/Prepare`:

- `awg-quick` shebang changed from `/bin/bash` to `/opt/bin/bash`.
- default config path changed from `/etc/amnezia/amneziawg` to `/opt/etc/amnezia/amneziawg`.

The resulting package depends on:

```text
libc, libssp, librt, libpthread, bash, ip-full
```
