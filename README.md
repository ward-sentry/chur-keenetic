# Chur Keenetic

<p align="center">
  <b>VPN manager for Keenetic routers with Entware</b>
</p>

<p align="center">
  <a href="#-русский">Русский</a> ·
  <a href="#-english">English</a>
</p>

<p align="center">
  <a href="https://ward-sentry.github.io/chur-keenetic/latest/aarch64-3.10/Packages"><img alt="latest aarch64" src="https://img.shields.io/badge/latest-aarch64--3.10-1f6feb"></a>
  <a href="https://ward-sentry.github.io/chur-keenetic/latest/mips-3.4/Packages"><img alt="latest mips" src="https://img.shields.io/badge/latest-mips--3.4-1f6feb"></a>
  <a href="https://ward-sentry.github.io/chur-keenetic/latest/mipsel-3.4/Packages"><img alt="latest mipsel" src="https://img.shields.io/badge/latest-mipsel--3.4-1f6feb"></a>
</p>

<p align="center">
  <a href="https://ward-sentry.github.io/chur-keenetic/latest/"><img alt="latest feed" src="https://img.shields.io/badge/feed-latest-147d4b"></a>
  <a href="https://ward-sentry.github.io/chur-keenetic/0_1_0/"><img alt="0.1.0 feed" src="https://img.shields.io/badge/release-0.1.0-5c6b78"></a>
</p>

---

## 🇷🇺 Русский

**Chur Keenetic** — это веб-менеджер VPN-интерфейсов для роутеров Keenetic с Entware.

Идея простая: вместо длинной ручной настройки в терминале вы устанавливаете один пакет, открываете веб-интерфейс на роутере и управляете VPN через браузер.

### Что уже умеет

- Устанавливается как opkg-пакет `chur-keenetic`.
- Запускает веб-интерфейс на роутере.
- Проверяет системное состояние и runtime-зависимости.
- Устанавливает AmneziaWG runtime только когда он нужен.
- Создает несколько AmneziaWG-интерфейсов: `opkgtun0`, `opkgtun1`, `opkgtun2`, ...
- Позволяет загрузить или заменить `.conf` файл.
- Позволяет менять отображаемое имя интерфейса и MTU.
- Удаляет созданные интерфейсы и свои конфиги.

### В планах

- OpenVPN.
- WireGuard.
- VLESS.
- Другие VPN/proxy runtime.

### Быстрая установка

1. Проверьте архитектуру роутера:

```sh
opkg print-architecture
```

2. Выберите feed:

| Архитектура в Entware | Команда |
| --- | --- |
| `aarch64-3.10` | `echo "src/gz chur https://ward-sentry.github.io/chur-keenetic/latest/aarch64-3.10" > /opt/etc/opkg/chur.conf` |
| `mips-3.4` | `echo "src/gz chur https://ward-sentry.github.io/chur-keenetic/latest/mips-3.4" > /opt/etc/opkg/chur.conf` |
| `mipsel-3.4` | `echo "src/gz chur https://ward-sentry.github.io/chur-keenetic/latest/mipsel-3.4" > /opt/etc/opkg/chur.conf` |

Если на роутере есть, например:

```text
arch all 100
arch aarch64-3.10 150
arch aarch64-3.10_kn 200
```

используйте `aarch64-3.10`.

3. Установите пакет:

```sh
mkdir -p /opt/etc/opkg
opkg update
opkg install chur-keenetic
/opt/etc/init.d/S99chur-keenetic start
```

4. Откройте веб-интерфейс:

```text
http://<ip-роутера>:8088/
```

Пример:

```text
http://192.168.88.1:8088/
```

### Обновление

```sh
opkg update
opkg upgrade chur-keenetic
/opt/etc/init.d/S99chur-keenetic restart
```

### Удаление

Команда удалит Chur-пакеты, созданные через Chur интерфейсы и конфиги:

```sh
/opt/etc/init.d/S99chur-keenetic remove
```

### Как работает AmneziaWG

`chur-keenetic` устанавливает только менеджер.

AmneziaWG ставится отдельно из веб-интерфейса, когда вы выбираете AmneziaWG и нажимаете установку runtime. Тогда будут установлены:

- `chur-amneziawg`
- `chur-amneziawg-go`
- `chur-amneziawg-tools`

После этого можно загрузить `.conf` файл, сохранить интерфейс, изменить его название, MTU или заменить конфиг.

### Версии feed

Для обычной установки используйте `latest`:

```text
https://ward-sentry.github.io/chur-keenetic/latest/<arch>
```

Для установки конкретного релиза используйте версионную папку:

```text
https://ward-sentry.github.io/chur-keenetic/0_1_0/<arch>
```

<details>
<summary><b>Команды разработки</b></summary>

```sh
make test
make build_all
make build_runtime_amneziawg_go_all
```

Сборка IPK для архитектур:

```sh
ENTWARE_TARGET_CONFIG=aarch64-3.10 make entware_build_ipk_volume
ENTWARE_TARGET_CONFIG=mips-3.4 make entware_build_ipk_volume
ENTWARE_TARGET_CONFIG=mipsel-3.4 make entware_build_ipk_volume
```

Генерация opkg feed:

```sh
OPKG_REPO_ARCH=aarch64-3.10 make opkg_repo
OPKG_REPO_ARCH=mips-3.4 make opkg_repo
OPKG_REPO_ARCH=mipsel-3.4 make opkg_repo
```

Копирование feed в GitHub Pages репозиторий:

```sh
make clone_release
```

</details>

---

## 🇬🇧 English

**Chur Keenetic** is a web manager for VPN interfaces on Keenetic routers with Entware.

The idea is simple: instead of doing a long manual terminal setup, install one package, open the router web UI, and manage VPN interfaces from the browser.

### Current Features

- Installs as the `chur-keenetic` opkg package.
- Runs a web UI on the router.
- Shows system and runtime diagnostics.
- Installs the AmneziaWG runtime only when needed.
- Creates multiple AmneziaWG interfaces: `opkgtun0`, `opkgtun1`, `opkgtun2`, ...
- Uploads or replaces `.conf` files.
- Edits display name and MTU.
- Removes Chur-managed interfaces and configs.

### Roadmap

- OpenVPN.
- WireGuard.
- VLESS.
- More VPN/proxy runtimes.

### Quick Install

1. Check your router architecture:

```sh
opkg print-architecture
```

2. Select the feed:

| Entware architecture | Command |
| --- | --- |
| `aarch64-3.10` | `echo "src/gz chur https://ward-sentry.github.io/chur-keenetic/latest/aarch64-3.10" > /opt/etc/opkg/chur.conf` |
| `mips-3.4` | `echo "src/gz chur https://ward-sentry.github.io/chur-keenetic/latest/mips-3.4" > /opt/etc/opkg/chur.conf` |
| `mipsel-3.4` | `echo "src/gz chur https://ward-sentry.github.io/chur-keenetic/latest/mipsel-3.4" > /opt/etc/opkg/chur.conf` |

If your router shows something like:

```text
arch all 100
arch aarch64-3.10 150
arch aarch64-3.10_kn 200
```

use `aarch64-3.10`.

3. Install the package:

```sh
mkdir -p /opt/etc/opkg
opkg update
opkg install chur-keenetic
/opt/etc/init.d/S99chur-keenetic start
```

4. Open the web UI:

```text
http://<router-ip>:8088/
```

Example:

```text
http://192.168.88.1:8088/
```

### Update

```sh
opkg update
opkg upgrade chur-keenetic
/opt/etc/init.d/S99chur-keenetic restart
```

### Remove

This command removes Chur packages, Chur-managed interfaces, and configs:

```sh
/opt/etc/init.d/S99chur-keenetic remove
```

### How AmneziaWG Works

`chur-keenetic` installs only the manager itself.

AmneziaWG is installed separately from the web UI when you select AmneziaWG and install the runtime. These packages will be installed:

- `chur-amneziawg`
- `chur-amneziawg-go`
- `chur-amneziawg-tools`

After that, upload a `.conf` file, save the interface, edit its display name, change MTU, or replace the config.

### Feed Versions

Use `latest` for normal installation:

```text
https://ward-sentry.github.io/chur-keenetic/latest/<arch>
```

Use a versioned folder to pin a specific release:

```text
https://ward-sentry.github.io/chur-keenetic/0_1_0/<arch>
```

<details>
<summary><b>Development Commands</b></summary>

```sh
make test
make build_all
make build_runtime_amneziawg_go_all
```

Build IPK packages:

```sh
ENTWARE_TARGET_CONFIG=aarch64-3.10 make entware_build_ipk_volume
ENTWARE_TARGET_CONFIG=mips-3.4 make entware_build_ipk_volume
ENTWARE_TARGET_CONFIG=mipsel-3.4 make entware_build_ipk_volume
```

Generate opkg feeds:

```sh
OPKG_REPO_ARCH=aarch64-3.10 make opkg_repo
OPKG_REPO_ARCH=mips-3.4 make opkg_repo
OPKG_REPO_ARCH=mipsel-3.4 make opkg_repo
```

Copy the feed into the GitHub Pages repository:

```sh
make clone_release
```

</details>
