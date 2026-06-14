# Upstream Runtime Sources

`chur-keenetic` is a manager. VPN runtime components are sourced from original upstream repositories and pinned by commit.

Pinned upstream refs live in `.env`.

## Submodules

```text
third_party/amneziawg-go
```

Initialize after clone:

```sh
git submodule update --init --recursive
```

`amneziawg-tools` is intentionally not required as a submodule for Chur builds.
The Entware package recipe downloads the official GitHub archive at the pinned
commit and verifies its SHA256.

## Pinned Refs

### amneziawg-go

Repository:

```text
https://github.com/amnezia-vpn/amneziawg-go
```

Pinned commit from `.env`:

```text
f4f4c999267437c3eb909e8d0e5278fb4596d9a7
```

Observed description:

```text
f4f4c99 fix: apply S4 transport padding to keepalive packets
```

Submodule status:

```text
f4f4c999267437c3eb909e8d0e5278fb4596d9a7 third_party/amneziawg-go (v0.2.9-36-gf4f4c99)
```

### amneziawg-tools

Repository:

```text
https://github.com/amnezia-vpn/amneziawg-tools
```

Pinned commit from `.env`:

```text
5d6179a6d0842e98dfb349c28cf1bd8e4b9d1079
```

Observed description:

```text
5d6179a Merge pull request #38 from amnezia-vpn/fix/android-i1-i5
```

Build source:

```text
GitHub archive, verified by Entware PKG_HASH
```

## Build Notes

`amneziawg-go`:

- Go project.
- Upstream `Makefile` runs `go build -v -o amneziawg-go`.
- First target for Ultra: `GOOS=linux GOARCH=arm64`.

`amneziawg-tools`:

- C project under `src/`.
- Upstream build command: `cd src && make`.
- Installs `wg` as `awg`.
- Installs `wg-quick/<platform>.bash` as `awg-quick` when `WITH_WGQUICK=yes`.
- Cross-building for Entware will require an appropriate C cross-compiler/toolchain.

Important upstream note for our architecture: for a real network manager, integrate with `awg` or the direct API instead of relying only on `awg-quick`.
