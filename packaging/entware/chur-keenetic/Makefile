include $(TOPDIR)/rules.mk

PKG_NAME:=chur-keenetic
CHUR_VERSION?=1.0.0
PKG_VERSION:=$(CHUR_VERSION)
PKG_RELEASE:=1
PKG_LICENSE:=Chur-Keenetic-Attribution-1.0

CHUR_MANAGER_BINARY?=/src/dist/chur_$(PKG_VERSION)_entware_aarch64-3.10

PKG_BUILD_DIR:=$(BUILD_DIR)/$(PKG_NAME)-$(PKG_VERSION)

include $(INCLUDE_DIR)/package.mk

define Package/chur-keenetic
  SECTION:=net
  CATEGORY:=Network
  SUBMENU:=VPN
  URL:=https://github.com/ward-sentry/chur-keenetic
  MAINTAINER:=Chur Keenetic
  TITLE:=Web manager for VPN runtimes on Keenetic Entware
endef

define Package/chur-keenetic/description
  Chur Keenetic web manager. Installs the manager binary and an Entware
  init script that starts the web UI on port 8088.
endef

define Build/Prepare
	mkdir -p $(PKG_BUILD_DIR)
endef

define Build/Configure
endef

define Build/Compile
	test -x "$(CHUR_MANAGER_BINARY)"
endef

define Package/chur-keenetic/install
	$(INSTALL_DIR) $(1)/opt/bin
	$(INSTALL_BIN) $(CHUR_MANAGER_BINARY) $(1)/opt/bin/chur-keenetic
	$(INSTALL_DIR) $(1)/opt/etc/init.d
	$(INSTALL_BIN) ./files/S99chur-keenetic $(1)/opt/etc/init.d/S99chur-keenetic
	$(INSTALL_DIR) $(1)/opt/etc/chur-keenetic
	$(INSTALL_DIR) $(1)/opt/var/log
	$(INSTALL_DIR) $(1)/opt/var/run
endef

define Package/chur-keenetic/postinst
#!/bin/sh
[ -n "$${IPKG_INSTROOT}" ] || /opt/etc/init.d/S99chur-keenetic start || true
endef

define Package/chur-keenetic/prerm
#!/bin/sh
[ -n "$${IPKG_INSTROOT}" ] || /opt/etc/init.d/S99chur-keenetic stop || true
endef

$(eval $(call BuildPackage,chur-keenetic))
