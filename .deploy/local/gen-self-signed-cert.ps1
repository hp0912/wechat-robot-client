Param(
  [string]$CertDir = "$(Split-Path -Parent $PSCommandPath)\secrets\nginx",
  [int]$Days = 3650,
  [string[]]$DnsNames = @(),
  [string[]]$IpAddresses = @()
)

$ErrorActionPreference = 'Stop'

New-Item -ItemType Directory -Force -Path $CertDir | Out-Null

$crtPath = Join-Path $CertDir 'tls.crt'
$keyPath = Join-Path $CertDir 'tls.key'

function Get-OpenSslCommand {
  $cmd = Get-Command openssl -ErrorAction SilentlyContinue
  if ($cmd) { return $cmd.Source }

  # Common Git-for-Windows path (best-effort)
  $gitOpenSsl = Join-Path $Env:ProgramFiles 'Git\usr\bin\openssl.exe'
  if (Test-Path $gitOpenSsl) { return $gitOpenSsl }

  throw "openssl not found. Install OpenSSL (or Git for Windows) or run the sh script under WSL/Git-Bash."
}

$openssl = Get-OpenSslCommand

$tmp = New-TemporaryFile
$confPath = "$($tmp.FullName).cnf"
Move-Item -Force $tmp.FullName $confPath

$confLines = @(
  '[ req ]',
  'default_bits       = 2048',
  'prompt             = no',
  'default_md         = sha256',
  'x509_extensions    = v3_req',
  'distinguished_name = dn',
  '',
  '[ dn ]',
  'C  = CN',
  'ST = Shanghai',
  'L  = Shanghai',
  'O  = wechat-robot',
  'CN = localhost',
  '',
  '[ v3_req ]',
  'subjectAltName = @alt_names',
  'keyUsage = keyEncipherment, dataEncipherment',
  'extendedKeyUsage = serverAuth',
  '',
  '[ alt_names ]',
  'DNS.1 = localhost',
  'IP.1  = 127.0.0.1'
)

$dnsIndex = 2
foreach ($dns in $DnsNames) {
  if ([string]::IsNullOrWhiteSpace($dns)) { continue }
  $confLines += ("DNS.{0} = {1}" -f $dnsIndex, $dns)
  $dnsIndex++
}

$ipIndex = 2
foreach ($ip in $IpAddresses) {
  if ([string]::IsNullOrWhiteSpace($ip)) { continue }
  $confLines += ("IP.{0}  = {1}" -f $ipIndex, $ip)
  $ipIndex++
}

$confLines | Set-Content -Encoding ascii $confPath

try {
  & $openssl req -x509 -nodes -days $Days -newkey rsa:2048 -keyout $keyPath -out $crtPath -config $confPath | Out-Null
}
finally {
  Remove-Item -Force -ErrorAction SilentlyContinue $confPath
}

Write-Host "Generated:"
Write-Host "  - $crtPath"
Write-Host "  - $keyPath"
Write-Host "Tip: add LAN IP with: .\gen-self-signed-cert.ps1 -IpAddresses <A_LAN_IP>"
