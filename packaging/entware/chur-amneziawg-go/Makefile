include $(TOPDIR)/rules.mk

PKG_NAME:=chur-amneziawg-go
CHUR_AMNEZIAWG_GO_VERSION?=f4f4c99
PKG_VERSION:=$(CHUR_AMNEZIAWG_GO_VERSION)
PKG_RELEASE:=1

CHUR_AMNEZIAWG_GO_BINARY?=/src/dist/amneziawg-go_$(PKG_VERSION)_entware_aarch64-3.10

PKG_BUILD_DIR:=$(BUILD_DIR)/$(PKG_NAME)-$(PKG_VERSION)

include $(INCLUDE_DIR)/package.mk

define Package/chur-amneziawg-go
  SECTION:=net
  CATEGORY:=Network
  SUBMENU:=VPN
  URL:=https://github.com/amnezia-vpn/amneziawg-go
  MAINTAINER:=Chur Keenetic
  TITLE:=Pinned AmneziaWG userspace implementation for Chur Keenetic
endef

define Package/chur-amneziawg-go/description
  AmneziaWG userspace implementation built from the official upstream
  amnezia-vpn/amneziawg-go repository pinned by Chur Keenetic.
endef

define Build/Prepare
	mkdir -p $(PKG_BUILD_DIR)
endef

define Build/Configure
endef

define Build/Compile
	test -x "$(CHUR_AMNEZIAWG_GO_BINARY)"
endef

define Package/chur-amneziawg-go/install
	$(INSTALL_DIR) $(1)/opt/bin
	$(INSTALL_BIN) $(CHUR_AMNEZIAWG_GO_BINARY) $(1)/opt/bin/amneziawg-go
endef

$(eval $(call BuildPackage,chur-amneziawg-go))
