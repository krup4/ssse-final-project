#!/bin/sh
set -eu

escape_js_string() {
  printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'
}

yandex_maps_api_key="$(escape_js_string "${YANDEX_MAPS_API_KEY:-}")"

cat > /usr/share/nginx/html/config.js <<EOF
window.__APP_CONFIG__ = {
  yandexMapsApiKey: "${yandex_maps_api_key}"
};
EOF
