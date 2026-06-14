#!/bin/sh
set -eu

ARCH="${1:-aarch64-3.10}"
IPK_DIR="${IPK_DIR:-dist/entware-ipk}"
REPO_ROOT="${REPO_ROOT:-dist/opkg-repo}"
REPO_DIR="$REPO_ROOT/$ARCH"

mkdir -p "$REPO_DIR"
rm -f "$REPO_DIR"/Packages "$REPO_DIR"/Packages.gz
find "$REPO_DIR" -maxdepth 1 -type f -name '*.ipk' -delete

for ipk in "$IPK_DIR"/*_"$ARCH".ipk; do
	[ -f "$ipk" ] || continue
	case "$(basename "$ipk")" in
		chur_*.ipk)
			continue
			;;
	esac
	cp "$ipk" "$REPO_DIR/"
done

for ipk in "$REPO_DIR"/*.ipk; do
	[ -f "$ipk" ] || continue
	file=$(basename "$ipk")
	size=$(wc -c < "$ipk" | tr -d ' ')
	sha256=$(shasum -a 256 "$ipk" | awk '{print $1}')
	md5sum=$(md5 -q "$ipk")

	tar -xOzf "$ipk" ./control.tar.gz | tar -xOzf - ./control |
		awk -v filename="$file" -v size="$size" -v sha256="$sha256" -v md5sum="$md5sum" '
			BEGIN { printed_filename = 0 }
			/^Filename:/ { next }
			/^Size:/ { next }
			/^SHA256sum:/ { next }
			/^MD5Sum:/ { next }
			{ print }
			END {
				print "Filename: " filename
				print "Size: " size
				print "SHA256sum: " sha256
				print "MD5Sum: " md5sum
				print ""
			}
		' >> "$REPO_DIR/Packages"
done

gzip -9c "$REPO_DIR/Packages" > "$REPO_DIR/Packages.gz"

echo "Generated $REPO_DIR/Packages.gz"
