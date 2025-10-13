#!/bin/sh
set -e

OWNER="bab-sh"
REPO="bab"
BINARY="bab"
PROJECT_NAME="bab"

usage() {
  this=$1
  cat <<EOF
$this: download ${PROJECT_NAME} binary from GitHub releases

Usage: $this [-b bindir] [-d] [tag]
  -b sets bindir or installation directory, Defaults to ~/.local/bin
  -d turns on debug logging
  [tag] is a tag from https://github.com/${OWNER}/${REPO}/releases
        If tag is missing, then the latest will be used.

Examples:
  # Install to ~/.local/bin (default, no sudo required)
  curl -sSfL https://raw.githubusercontent.com/${OWNER}/${REPO}/main/install.sh | sh

  # Install to /usr/local/bin (system-wide, requires sudo)
  curl -sSfL https://raw.githubusercontent.com/${OWNER}/${REPO}/main/install.sh | sudo sh -s -- -b /usr/local/bin

  # Install specific version
  curl -sSfL https://raw.githubusercontent.com/${OWNER}/${REPO}/main/install.sh | sh -s -- v0.0.7

  # Install to custom directory
  curl -sSfL https://raw.githubusercontent.com/${OWNER}/${REPO}/main/install.sh | sh -s -- -b /custom/path

EOF
  exit 2
}

get_default_bindir() {
  if [ -d "$HOME/.local/bin" ]; then
    echo "$HOME/.local/bin"
  elif [ -w "/usr/local/bin" ] 2>/dev/null; then
    echo "/usr/local/bin"
  else
    echo "$HOME/.local/bin"
  fi
}

parse_args() {
  BINDIR=${BINDIR:-$(get_default_bindir)}
  while getopts "b:dh?x" arg; do
    case "$arg" in
      b) BINDIR="$OPTARG" ;;
      d) log_set_priority 10 ;;
      h | \?) usage "$0" ;;
      x) set -x ;;
    esac
  done
  shift $((OPTIND - 1))
  TAG=$1
}

check_path() {
  case ":${PATH}:" in
    *":${BINDIR}:"*) return 0 ;;
    *) return 1 ;;
  esac
}

execute() {
  tmpdir=$(mktemp -d)
  log_debug "downloading files into ${tmpdir}"
  http_download "${tmpdir}/${TARBALL}" "${TARBALL_URL}"
  http_download "${tmpdir}/${CHECKSUM}" "${CHECKSUM_URL}"
  hash_sha256_verify "${tmpdir}/${TARBALL}" "${tmpdir}/${CHECKSUM}"
  srcdir="${tmpdir}"
  (cd "${tmpdir}" && untar "${TARBALL}")
  test ! -d "${BINDIR}" && install -d "${BINDIR}"
  for binexe in $BINARIES; do
    if [ "$OS" = "windows" ]; then
      binexe="${binexe}.exe"
    fi
    install "${srcdir}/${binexe}" "${BINDIR}/"
    log_info "installed ${BINDIR}/${binexe}"
  done
  rm -rf "${tmpdir}"

  if ! check_path; then
    log_info ""
    log_info "==============================================="
    log_info "${BINDIR} is not in your PATH."
    log_info "Add it to your PATH by running one of these:"
    log_info ""
    if [ -f "$HOME/.bashrc" ]; then
      log_info "  echo 'export PATH=\"${BINDIR}:\$PATH\"' >> ~/.bashrc"
      log_info "  source ~/.bashrc"
    fi
    if [ -f "$HOME/.zshrc" ]; then
      log_info "  echo 'export PATH=\"${BINDIR}:\$PATH\"' >> ~/.zshrc"
      log_info "  source ~/.zshrc"
    fi
    if [ -f "$HOME/.config/fish/config.fish" ]; then
      log_info "  fish_add_path ${BINDIR}"
    fi
    log_info "==============================================="
  fi
}

is_command() {
  command -v "$1" >/dev/null
}

echoerr() {
  echo "$@" 1>&2
}

_logp=6
log_set_priority() {
  _logp="$1"
}

log_priority() {
  if test -z "$1"; then
    echo "$_logp"
    return
  fi
  [ "$1" -le "$_logp" ]
}

log_debug() {
  log_priority 7 || return 0
  echoerr "$@"
}

log_info() {
  log_priority 6 || return 0
  echoerr "$@"
}

log_err() {
  log_priority 3 || return 0
  echoerr "$@"
}

log_crit() {
  log_priority 2 || return 0
  echoerr "$@"
}

uname_os() {
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  case "$os" in
    cygwin_nt*) os="windows" ;;
    mingw*) os="windows" ;;
    msys_nt*) os="windows" ;;
  esac
  echo "$os"
}

uname_arch() {
  arch=$(uname -m)
  case $arch in
    x86_64) arch="amd64" ;;
    x86) arch="386" ;;
    i686) arch="386" ;;
    i386) arch="386" ;;
    aarch64) arch="arm64" ;;
    armv5*) arch="armv5" ;;
    armv6*) arch="armv6" ;;
    armv7*) arch="armv7" ;;
  esac
  echo "${arch}"
}

uname_os_check() {
  os=$(uname_os)
  case "$os" in
    darwin) return 0 ;;
    dragonfly) return 0 ;;
    freebsd) return 0 ;;
    linux) return 0 ;;
    android) return 0 ;;
    nacl) return 0 ;;
    netbsd) return 0 ;;
    openbsd) return 0 ;;
    plan9) return 0 ;;
    solaris) return 0 ;;
    windows) return 0 ;;
  esac
  log_crit "uname_os_check '$(uname -s)' got converted to '$os' which is not supported"
  return 1
}

uname_arch_check() {
  arch=$(uname_arch)
  case "$arch" in
    386) return 0 ;;
    amd64) return 0 ;;
    arm64) return 0 ;;
    armv5) return 0 ;;
    armv6) return 0 ;;
    armv7) return 0 ;;
    ppc64) return 0 ;;
    ppc64le) return 0 ;;
    mips64) return 0 ;;
    mips64le) return 0 ;;
    s390x) return 0 ;;
    amd64p32) return 0 ;;
  esac
  log_crit "uname_arch_check '$(uname -m)' got converted to '$arch' which is not supported"
  return 1
}

untar() {
  tarball=$1
  case "${tarball}" in
    *.tar.gz | *.tgz) tar --no-same-owner -xzf "${tarball}" ;;
    *.tar) tar --no-same-owner -xf "${tarball}" ;;
    *.zip) unzip -q "${tarball}" ;;
    *)
      log_err "untar: unknown archive format for ${tarball}"
      return 1
      ;;
  esac
}

http_download_curl() {
  local_file=$1
  source_url=$2
  header=$3
  if [ -z "$header" ]; then
    code=$(curl -w '%{http_code}' -sL -o "$local_file" "$source_url")
  else
    code=$(curl -w '%{http_code}' -sL -H "$header" -o "$local_file" "$source_url")
  fi
  if [ "$code" != "200" ]; then
    log_err "http_download_curl received HTTP status $code"
    return 1
  fi
  return 0
}

http_download_wget() {
  local_file=$1
  source_url=$2
  header=$3
  if [ -z "$header" ]; then
    wget -q -O "$local_file" "$source_url"
  else
    wget -q --header "$header" -O "$local_file" "$source_url"
  fi
}

http_download() {
  log_debug "http_download $2"
  if is_command curl; then
    http_download_curl "$@"
    return
  elif is_command wget; then
    http_download_wget "$@"
    return
  fi
  log_crit "http_download unable to find wget or curl"
  return 1
}

http_copy() {
  tmp=$(mktemp)
  http_download "${tmp}" "$1" "$2" || return 1
  body=$(cat "$tmp")
  rm -f "${tmp}"
  echo "$body"
}

github_release() {
  owner_repo=$1
  version=$2
  test -z "$version" && version="latest"
  giturl="https://github.com/${owner_repo}/releases/${version}"
  json=$(http_copy "$giturl" "Accept:application/json")
  test -z "$json" && return 1
  version=$(echo "$json" | tr -s '\n' ' ' | sed 's/.*"tag_name":"//' | sed 's/".*//')
  test -z "$version" && return 1
  echo "$version"
}

hash_sha256() {
  TARGET=${1:-/dev/stdin}
  if is_command gsha256sum; then
    hash=$(gsha256sum "$TARGET") || return 1
    echo "$hash" | cut -d ' ' -f 1
  elif is_command sha256sum; then
    hash=$(sha256sum "$TARGET") || return 1
    echo "$hash" | cut -d ' ' -f 1
  elif is_command shasum; then
    hash=$(shasum -a 256 "$TARGET" 2>/dev/null) || return 1
    echo "$hash" | cut -d ' ' -f 1
  elif is_command openssl; then
    hash=$(openssl dgst -sha256 "$TARGET") || return 1
    echo "$hash" | cut -d ' ' -f a
  else
    log_crit "hash_sha256 unable to find command to compute sha-256 hash"
    return 1
  fi
}

hash_sha256_verify() {
  TARGET=$1
  checksums=$2
  if [ -z "$checksums" ]; then
    log_err "hash_sha256_verify checksum file not specified"
    return 1
  fi
  BASENAME=${TARGET##*/}
  want=$(grep "${BASENAME}" "${checksums}" 2>/dev/null | tr '\t' ' ' | cut -d ' ' -f 1)
  if [ -z "$want" ]; then
    log_err "hash_sha256_verify unable to find checksum for '${TARGET}' in '${checksums}'"
    return 1
  fi
  got=$(hash_sha256 "$TARGET")
  if [ "$want" != "$got" ]; then
    log_err "hash_sha256_verify checksum for '$TARGET' did not verify ${want} vs $got"
    return 1
  fi
}

tag_to_version() {
  if [ -z "${TAG}" ]; then
    log_info "checking GitHub for latest tag"
  else
    log_info "checking GitHub for tag '${TAG}'"
  fi
  REALTAG=$(github_release "$OWNER/$REPO" "${TAG}") && true
  if test -z "$REALTAG"; then
    log_crit "unable to find '${TAG}' - use 'latest' or see https://github.com/${OWNER}/${REPO}/releases for details"
    exit 1
  fi
  TAG="$REALTAG"
  VERSION=${TAG#v}
}

adjust_format() {
  case ${OS} in
    windows) FORMAT=zip ;;
  esac
  true
}

adjust_os() {
  case ${OS} in
    darwin) ADJUSTED_OS="macOS" ;;
    linux) ADJUSTED_OS="Linux" ;;
    windows) ADJUSTED_OS="Windows" ;;
    freebsd) ADJUSTED_OS="Freebsd" ;;
    *) ADJUSTED_OS=$(echo "${OS}" | sed 's/./\u&/') ;;
  esac
}

adjust_arch() {
  case ${ARCH} in
    amd64) ADJUSTED_ARCH="x86_64" ;;
    386) ADJUSTED_ARCH="i386" ;;
    arm64) ADJUSTED_ARCH="arm64" ;;
    armv7) ADJUSTED_ARCH="armv7" ;;
    armv6) ADJUSTED_ARCH="armv6" ;;
    armv5) ADJUSTED_ARCH="armv5" ;;
    *) ADJUSTED_ARCH="${ARCH}" ;;
  esac
}

FORMAT=tar.gz
OS=$(uname_os)
ARCH=$(uname_arch)
PREFIX="$OWNER/$REPO"

uname_os_check "$OS"
uname_arch_check "$ARCH"

parse_args "$@"

BINARIES="${BINARY}"
tag_to_version
adjust_format
adjust_os
adjust_arch

log_info "found version: ${VERSION} for ${TAG}/${OS}/${ARCH}"

NAME=${PROJECT_NAME}_${VERSION}_${ADJUSTED_OS}_${ADJUSTED_ARCH}
TARBALL=${NAME}.${FORMAT}
TARBALL_URL=https://github.com/${OWNER}/${REPO}/releases/download/${TAG}/${TARBALL}
CHECKSUM=checksums.txt
CHECKSUM_URL=https://github.com/${OWNER}/${REPO}/releases/download/${TAG}/${CHECKSUM}

execute
