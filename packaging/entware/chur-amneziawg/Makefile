include $(TOPDIR)/rules.mk

PKG_NAME:=chur-amneziawg
CHUR_VERSION?=1.0.0
PKG_VERSION:=$(CHUR_VERSION)
PKG_RELEASE:=1
PKG_LICENSE:=Chur-Keenetic-Attribution-1.0

PKG_BUILD_DIR:=$(BUILD_DIR)/$(PKG_NAME)-$(PKG_VERSION)

include $(INCLUDE_DIR)/package.mk

define Package/chur-amneziawg
  SECTION:=net
  CATEGORY:=Network
  SUBMENU:=VPN
  URL:=https://github.com/ward-sentry/chur-keenetic
  MAINTAINER:=Chur Keenetic
  TITLE:=AmneziaWG runtime provider for Chur Keenetic
  DEPENDS:=+chur-amneziawg-go +chur-amneziawg-tools
endef

define Package/chur-amneziawg/description
  Meta package that installs all AmneziaWG runtime dependencies required by
  Chur Keenetic.
endef

define Build/Prepare
	mkdir -p $(PKG_BUILD_DIR)
endef

define Build/Configure
endef

define Build/Compile
endef

define Package/chur-amneziawg/install
	$(INSTALL_DIR) $(1)/opt/etc/chur-keenetic/providers
endef

$(eval $(call BuildPackage,chur-amneziawg))
