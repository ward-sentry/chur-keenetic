GO ?= go
-include .env

GOCACHE ?= $(CURDIR)/.cache/go-build
GOMODCACHE ?= $(CURDIR)/.cache/go-mod
DIST_DIR ?= $(CURDIR)/dist
ENTWARE_DOCKER_IMAGE ?= chur-entware-build:20.04
OPKG_REPO_ARCH ?= aarch64-3.10
OPKG_REPO_PORT ?= 8090
CHUR_PACKAGES := ./cmd/... ./internal/...
VERSION := $(strip $(shell cat VERSION))
RELEASE_SITE_DIR ?= /Users/sf/Project/ward-sentry.github.io
RELEASE_VERSION_DIR := $(subst .,_,$(VERSION))
RELEASE_DIR ?= $(RELEASE_SITE_DIR)/chur-keenetic/$(RELEASE_VERSION_DIR)
RELEASE_LATEST_DIR ?= $(RELEASE_SITE_DIR)/chur-keenetic/latest
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
AMNEZIAWG_GO_DIR := $(CURDIR)/third_party/amneziawg-go
AMNEZIAWG_GO_COMMIT ?= f4f4c999267437c3eb909e8d0e5278fb4596d9a7
AMNEZIAWG_GO_SHORT_COMMIT := $(shell printf '%.7s' "$(AMNEZIAWG_GO_COMMIT)")
LDFLAGS := -s -w -X github.com/ward-sentry/chur-keenetic/internal/buildinfo.Version=$(VERSION) -X github.com/ward-sentry/chur-keenetic/internal/buildinfo.Commit=$(COMMIT)

.PHONY: test build_all build_all_with_ipk build_runtime_amneziawg_go_all build_entware_aarch64-3.10 build_entware_aarch64-3.10_kn build_entware_mips-3.4 build_entware_mipsel-3.4 build_entware_mips-3.4_kn build_entware_mipsel-3.4_kn build_runtime_amneziawg_go_entware_aarch64-3.10 build_runtime_amneziawg_go_entware_aarch64-3.10_kn build_runtime_amneziawg_go_entware_mips-3.4 build_runtime_amneziawg_go_entware_mips-3.4_kn build_runtime_amneziawg_go_entware_mipsel-3.4 build_runtime_amneziawg_go_entware_mipsel-3.4_kn entware_docker_image entware_build_ipk_volume opkg_repo opkg_repo_serve clone_release clean

test:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)"
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" $(GO) test $(CHUR_PACKAGES)

build_all: build_entware_aarch64-3.10 build_entware_aarch64-3.10_kn build_entware_mips-3.4 build_entware_mips-3.4_kn build_entware_mipsel-3.4 build_entware_mipsel-3.4_kn

build_all_with_ipk: build_all build_runtime_amneziawg_go_entware_aarch64-3.10 entware_build_ipk_volume

build_runtime_amneziawg_go_all: build_runtime_amneziawg_go_entware_aarch64-3.10 build_runtime_amneziawg_go_entware_aarch64-3.10_kn build_runtime_amneziawg_go_entware_mips-3.4 build_runtime_amneziawg_go_entware_mips-3.4_kn build_runtime_amneziawg_go_entware_mipsel-3.4 build_runtime_amneziawg_go_entware_mipsel-3.4_kn

opkg_repo:
	sh scripts/opkg-make-repo.sh "$(OPKG_REPO_ARCH)"

opkg_repo_serve: opkg_repo
	python3 -m http.server "$(OPKG_REPO_PORT)" --bind 0.0.0.0 --directory "$(DIST_DIR)/opkg-repo"

clone_release:
	test -d "$(RELEASE_SITE_DIR)/.git"
	test -d "$(DIST_DIR)/opkg-repo"
	mkdir -p "$(RELEASE_DIR)"
	mkdir -p "$(RELEASE_LATEST_DIR)"
	rsync -a --delete "$(DIST_DIR)/opkg-repo/" "$(RELEASE_DIR)/"
	rsync -a --delete "$(DIST_DIR)/opkg-repo/" "$(RELEASE_LATEST_DIR)/"
	@printf '%s\n' "Release feed copied to $(RELEASE_DIR)"
	@printf '%s\n' "Latest feed copied to $(RELEASE_LATEST_DIR)"

build_entware_aarch64-3.10:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)" "$(DIST_DIR)"
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o "$(DIST_DIR)/chur_$(VERSION)_entware_aarch64-3.10" ./cmd/chur-keenetic

build_entware_aarch64-3.10_kn:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)" "$(DIST_DIR)"
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o "$(DIST_DIR)/chur_$(VERSION)_entware_aarch64-3.10_kn" ./cmd/chur-keenetic

build_entware_mips-3.4:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)" "$(DIST_DIR)"
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" GOOS=linux GOARCH=mips GOMIPS=softfloat CGO_ENABLED=0 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o "$(DIST_DIR)/chur_$(VERSION)_entware_mips-3.4" ./cmd/chur-keenetic

build_entware_mips-3.4_kn:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)" "$(DIST_DIR)"
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" GOOS=linux GOARCH=mips GOMIPS=softfloat CGO_ENABLED=0 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o "$(DIST_DIR)/chur_$(VERSION)_entware_mips-3.4_kn" ./cmd/chur-keenetic

build_entware_mipsel-3.4:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)" "$(DIST_DIR)"
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" GOOS=linux GOARCH=mipsle GOMIPS=softfloat CGO_ENABLED=0 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o "$(DIST_DIR)/chur_$(VERSION)_entware_mipsel-3.4" ./cmd/chur-keenetic

build_entware_mipsel-3.4_kn:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)" "$(DIST_DIR)"
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" GOOS=linux GOARCH=mipsle GOMIPS=softfloat CGO_ENABLED=0 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o "$(DIST_DIR)/chur_$(VERSION)_entware_mipsel-3.4_kn" ./cmd/chur-keenetic

build_runtime_amneziawg_go_entware_aarch64-3.10_kn:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)" "$(DIST_DIR)"
	git -C "$(AMNEZIAWG_GO_DIR)" checkout --detach "$(AMNEZIAWG_GO_COMMIT)"
	cd "$(AMNEZIAWG_GO_DIR)" && GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GO) build -trimpath -ldflags="-s -w" -o "$(DIST_DIR)/amneziawg-go_$(AMNEZIAWG_GO_SHORT_COMMIT)_entware_aarch64-3.10_kn"

build_runtime_amneziawg_go_entware_aarch64-3.10:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)" "$(DIST_DIR)"
	git -C "$(AMNEZIAWG_GO_DIR)" checkout --detach "$(AMNEZIAWG_GO_COMMIT)"
	cd "$(AMNEZIAWG_GO_DIR)" && GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GO) build -trimpath -ldflags="-s -w" -o "$(DIST_DIR)/amneziawg-go_$(AMNEZIAWG_GO_SHORT_COMMIT)_entware_aarch64-3.10"

build_runtime_amneziawg_go_entware_mips-3.4:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)" "$(DIST_DIR)"
	git -C "$(AMNEZIAWG_GO_DIR)" checkout --detach "$(AMNEZIAWG_GO_COMMIT)"
	cd "$(AMNEZIAWG_GO_DIR)" && GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" GOOS=linux GOARCH=mips GOMIPS=softfloat CGO_ENABLED=0 $(GO) build -trimpath -ldflags="-s -w" -o "$(DIST_DIR)/amneziawg-go_$(AMNEZIAWG_GO_SHORT_COMMIT)_entware_mips-3.4"

build_runtime_amneziawg_go_entware_mips-3.4_kn:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)" "$(DIST_DIR)"
	git -C "$(AMNEZIAWG_GO_DIR)" checkout --detach "$(AMNEZIAWG_GO_COMMIT)"
	cd "$(AMNEZIAWG_GO_DIR)" && GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" GOOS=linux GOARCH=mips GOMIPS=softfloat CGO_ENABLED=0 $(GO) build -trimpath -ldflags="-s -w" -o "$(DIST_DIR)/amneziawg-go_$(AMNEZIAWG_GO_SHORT_COMMIT)_entware_mips-3.4_kn"

build_runtime_amneziawg_go_entware_mipsel-3.4:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)" "$(DIST_DIR)"
	git -C "$(AMNEZIAWG_GO_DIR)" checkout --detach "$(AMNEZIAWG_GO_COMMIT)"
	cd "$(AMNEZIAWG_GO_DIR)" && GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" GOOS=linux GOARCH=mipsle GOMIPS=softfloat CGO_ENABLED=0 $(GO) build -trimpath -ldflags="-s -w" -o "$(DIST_DIR)/amneziawg-go_$(AMNEZIAWG_GO_SHORT_COMMIT)_entware_mipsel-3.4"

build_runtime_amneziawg_go_entware_mipsel-3.4_kn:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)" "$(DIST_DIR)"
	git -C "$(AMNEZIAWG_GO_DIR)" checkout --detach "$(AMNEZIAWG_GO_COMMIT)"
	cd "$(AMNEZIAWG_GO_DIR)" && GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" GOOS=linux GOARCH=mipsle GOMIPS=softfloat CGO_ENABLED=0 $(GO) build -trimpath -ldflags="-s -w" -o "$(DIST_DIR)/amneziawg-go_$(AMNEZIAWG_GO_SHORT_COMMIT)_entware_mipsel-3.4_kn"

entware_docker_image:
	docker build -f docker/entware-build/Dockerfile -t "$(ENTWARE_DOCKER_IMAGE)" .

entware_build_ipk_volume:
	ENTWARE_DOCKER_IMAGE="$(ENTWARE_DOCKER_IMAGE)" sh scripts/entware-build-volume.sh

clean:
	rm -rf "$(DIST_DIR)"
