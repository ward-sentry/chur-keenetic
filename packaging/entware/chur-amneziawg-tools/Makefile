include $(TOPDIR)/rules.mk

PKG_NAME:=chur-amneziawg-tools
PKG_VERSION:=1.0.20260223
PKG_RELEASE:=2

CHUR_AMNEZIAWG_TOOLS_COMMIT?=5d6179a6d0842e98dfb349c28cf1bd8e4b9d1079
CHUR_AMNEZIAWG_TOOLS_SHA256?=e79a3c7f2def315d052a3648b49058a268c4b63cdb5e082b696d2a4a0a2367f0

PKG_SOURCE_VERSION:=$(CHUR_AMNEZIAWG_TOOLS_COMMIT)
PKG_SOURCE:=$(PKG_SOURCE_VERSION).tar.gz
PKG_SOURCE_URL:=https://github.com/amnezia-vpn/amneziawg-tools/archive
PKG_HASH:=$(CHUR_AMNEZIAWG_TOOLS_SHA256)
PKG_BUILD_DIR:=$(BUILD_DIR)/amneziawg-tools-$(PKG_SOURCE_VERSION)

PKG_LICENSE:=GPL-2.0
PKG_LICENSE_FILES:=COPYING

PKG_BUILD_PARALLEL:=1

include $(INCLUDE_DIR)/package.mk

define Package/chur-amneziawg-tools
  SECTION:=net
  CATEGORY:=Network
  SUBMENU:=VPN
  URL:=https://github.com/amnezia-vpn/amneziawg-tools
  MAINTAINER:=Chur Keenetic
  TITLE:=Pinned AmneziaWG userspace tools for Chur Keenetic
  DEPENDS:=+bash +ip-full +coreutils-stat
endef

define Package/chur-amneziawg-tools/description
  AmneziaWG userspace control tools built from the official upstream
  amnezia-vpn/amneziawg-tools repository pinned by Chur Keenetic.
  This package installs awg and awg-quick into /opt/bin.
endef

define Build/Prepare
		$(call Build/Prepare/Default)
		$(SED) '1s|^#!/bin/bash|#!/opt/bin/bash|' $(PKG_BUILD_DIR)/src/wg-quick/linux.bash
		$(SED) 's|/etc/amnezia/amneziawg|/opt/etc/amnezia/amneziawg|g' $(PKG_BUILD_DIR)/src/wg-quick/linux.bash
endef

define Build/Configure
endef

define Build/Compile
		CFLAGS="$(TARGET_CFLAGS)" \
		LDFLAGS="$(TARGET_LDFLAGS)" \
		$(MAKE) -C $(PKG_BUILD_DIR)/src \
			CC="$(TARGET_CC)" \
			PLATFORM=linux \
			wg
endef

define Package/chur-amneziawg-tools/install
	$(INSTALL_DIR) $(1)/opt/bin
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/src/wg $(1)/opt/bin/awg
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/src/wg-quick/linux.bash $(1)/opt/bin/awg-quick
	$(INSTALL_DIR) $(1)/opt/etc/amnezia/amneziawg
	$(INSTALL_DIR) $(1)/opt/var/run/amneziawg
endef

$(eval $(call BuildPackage,chur-amneziawg-tools))
