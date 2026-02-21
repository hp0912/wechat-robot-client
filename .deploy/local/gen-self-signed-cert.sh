#!/usr/bin/env sh
set -eu

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
CERT_DIR="$SCRIPT_DIR/secrets/nginx"

EXTRA_DNS_NAMES=""
EXTRA_IPS=""

usage() {
  cat <<'EOF'
Usage:
  gen-self-signed-cert.sh [--dns <name>]... [--ip <addr>]...

Examples:
  # default SAN: localhost + 127.0.0.1
  ./gen-self-signed-cert.sh

  # add LAN IP for access from other machines
  ./gen-self-signed-cert.sh --ip 192.168.1.10

  # add both DNS and IP
  ./gen-self-signed-cert.sh --dns wechat.local --ip 192.168.1.10
EOF
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --dns)
      [ "$#" -ge 2 ] || { echo "Missing value for --dns" >&2; usage >&2; exit 2; }
      EXTRA_DNS_NAMES="${EXTRA_DNS_NAMES} $2"
      shift 2
      ;;
    --ip)
      [ "$#" -ge 2 ] || { echo "Missing value for --ip" >&2; usage >&2; exit 2; }
      EXTRA_IPS="${EXTRA_IPS} $2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

mkdir -p "$CERT_DIR"

TMP_CONF=""

# mktemp compatibility (macOS/BSD, GNU coreutils, busybox)
if TMP_CONF="$(mktemp "${TMPDIR:-/tmp}/nginx-selfsigned.XXXXXX.cnf" 2>/dev/null)"; then
  :
elif TMP_CONF="$(mktemp -t nginx-selfsigned-XXXXXX.cnf 2>/dev/null)"; then
  :
else
  TMP_CONF="${TMPDIR:-/tmp}/nginx-selfsigned.$$.cnf"
  : >"$TMP_CONF"
fi

trap 'rm -f "$TMP_CONF"' EXIT

cat >"$TMP_CONF" <<'EOF'
[ req ]
default_bits       = 2048
prompt             = no
default_md         = sha256
x509_extensions    = v3_req
distinguished_name = dn

[ dn ]
C  = CN
ST = Shanghai
L  = Shanghai
O  = wechat-robot
CN = localhost

[ v3_req ]
subjectAltName = @alt_names
basicConstraints = CA:FALSE
keyUsage = critical, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
EOF

{
  echo ""
  echo "[ alt_names ]"
  dns_i=1
  echo "DNS.${dns_i} = localhost"
  for name in $EXTRA_DNS_NAMES; do
    [ -n "$name" ] || continue
    dns_i=$((dns_i + 1))
    echo "DNS.${dns_i} = ${name}"
  done

  ip_i=1
  echo "IP.${ip_i}  = 127.0.0.1"
  for ip in $EXTRA_IPS; do
    [ -n "$ip" ] || continue
    ip_i=$((ip_i + 1))
    echo "IP.${ip_i}  = ${ip}"
  done
} >>"$TMP_CONF"

openssl req \
  -x509 -nodes -days 3650 \
  -newkey rsa:2048 \
  -keyout "$CERT_DIR/tls.key" \
  -out "$CERT_DIR/tls.crt" \
  -config "$TMP_CONF" -extensions v3_req >/dev/null 2>&1

chmod 600 "$CERT_DIR/tls.key" || true
chmod 644 "$CERT_DIR/tls.crt" || true

echo "Generated:"
echo "  - $CERT_DIR/tls.crt"
echo "  - $CERT_DIR/tls.key"
echo "Tip: add LAN IP with: ./gen-self-signed-cert.sh --ip <A_LAN_IP>"