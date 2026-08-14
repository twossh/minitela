#!/usr/bin/env bash
set -Eeuo pipefail

ROOT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)"
APPIMAGE_DIR="${ROOT_DIR}/packaging/appimage"
DIST_DIR="${ROOT_DIR}/dist"

ARCH="$(uname -m)"
[[ "${ARCH}" == "x86_64" ]] || {
    echo "ERRO: este build v1 suporta apenas x86_64 (detectado: ${ARCH})." >&2
    exit 1
}

LINUXDEPLOY_VERSION="1-alpha-20251107-1"
APPIMAGETOOL_VERSION="1.9.1"

TOOLS_DIR="${XDG_CACHE_HOME:-${HOME}/.cache}/minitela/appimage-tools"
LINUXDEPLOY="${TOOLS_DIR}/linuxdeploy-x86_64.AppImage"
APPIMAGETOOL="${TOOLS_DIR}/appimagetool-x86_64.AppImage"

LINUXDEPLOY_URL="https://github.com/linuxdeploy/linuxdeploy/releases/download/${LINUXDEPLOY_VERSION}/linuxdeploy-x86_64.AppImage"
APPIMAGETOOL_URL="https://github.com/AppImage/appimagetool/releases/download/${APPIMAGETOOL_VERSION}/appimagetool-x86_64.AppImage"

WORK_DIR="$(mktemp -d "${TMPDIR:-/tmp}/minitela-appimage.XXXXXX")"
APPDIR="${WORK_DIR}/MiniTela.AppDir"
BUILD_GUI="${WORK_DIR}/minitela-gui"
BUILD_BACKEND="${WORK_DIR}/minitela"
BUILD_CTL="${WORK_DIR}/MiniTelaCtl"

cleanup() {
    rm -rf -- "${WORK_DIR}"
}
trap cleanup EXIT

log() {
    printf '\n==> %s\n' "$*"
}

die() {
    printf '\nERRO: %s\n' "$*" >&2
    exit 1
}

download_tool() {
    local url="$1"
    local output="$2"

    if [[ -x "${output}" ]]; then
        return
    fi

    mkdir -p "$(dirname -- "${output}")"

    log "Baixando $(basename -- "${output}")"

    if command -v curl >/dev/null 2>&1; then
        curl \
            --fail \
            --location \
            --retry 3 \
            --output "${output}" \
            "${url}"
    elif command -v wget >/dev/null 2>&1; then
        wget -O "${output}" "${url}"
    else
        die "curl ou wget é necessário para baixar as ferramentas AppImage"
    fi

    chmod +x "${output}"
}

command -v go >/dev/null 2>&1 || die "Go não encontrado"
command -v file >/dev/null 2>&1 || die "file não encontrado"

[[ -f "${ROOT_DIR}/go.mod" ]] || die "go.mod não encontrado"
[[ -x "${APPIMAGE_DIR}/AppRun" ]] || die "AppRun ausente ou sem permissão de execução"
[[ -f "${APPIMAGE_DIR}/minitela.desktop" ]] || die "minitela.desktop não encontrado"
[[ -f "${APPIMAGE_DIR}/minitela.svg" ]] || die "minitela.svg não encontrado"
[[ -f "${APPIMAGE_DIR}/minitela.service" ]] || die "minitela.service não encontrado"
[[ -f "${ROOT_DIR}/packaging/udev/99-minitela.rules" ]] || die "regra udev não encontrada"

download_tool "${LINUXDEPLOY_URL}" "${LINUXDEPLOY}"
download_tool "${APPIMAGETOOL_URL}" "${APPIMAGETOOL}"

log "Executando testes"
(
    cd "${ROOT_DIR}"
    go test ./...
)

log "Compilando MiniTela GUI, backend e controle em modo release"
(
    cd "${ROOT_DIR}"

    go build \
        -trimpath \
        -ldflags="-s -w" \
        -o "${BUILD_GUI}" \
        ./cmd/minitela-gui

    CGO_ENABLED=0 go build \
        -trimpath \
        -ldflags="-s -w" \
        -o "${BUILD_BACKEND}" \
        ./cmd/minitela

    CGO_ENABLED=0 go build \
        -trimpath \
        -ldflags="-s -w" \
        -o "${BUILD_CTL}" \
        ./cmd/minitela-ctl
)

file "${BUILD_GUI}"
file "${BUILD_BACKEND}"
file "${BUILD_CTL}"

rm -rf -- "${APPDIR}"
mkdir -p \
    "${APPDIR}/usr/bin" \
    "${APPDIR}/usr/share/minitela/systemd" \
    "${APPDIR}/usr/share/minitela/udev"

install -m 0755 "${BUILD_GUI}" "${APPDIR}/usr/bin/minitela-gui"
install -m 0644 \
    "${APPIMAGE_DIR}/minitela.service" \
    "${APPDIR}/usr/share/minitela/systemd/minitela.service"
install -m 0644 \
    "${ROOT_DIR}/packaging/udev/99-minitela.rules" \
    "${APPDIR}/usr/share/minitela/udev/99-minitela.rules"

# Build local: inclui os assets vendor se existirem.
# Eles estão ignorados pelo Git e NÃO devem ser publicados sem uma
# verificação específica dos direitos de redistribuição.
if [[ "${MINITELA_INCLUDE_VENDOR:-1}" == "1" ]]; then
    if [[ -d "${ROOT_DIR}/assets/vendor" ]]; then
        log "Incluindo assets/vendor no AppImage local"
        mkdir -p "${APPDIR}/usr/share/minitela/vendor"
        cp -a \
            "${ROOT_DIR}/assets/vendor/." \
            "${APPDIR}/usr/share/minitela/vendor/"
    else
        echo "AVISO: assets/vendor não encontrado; AppImage será criado sem galeria/template ACF." >&2
    fi
else
    log "Build sem assets vendor (MINITELA_INCLUDE_VENDOR=0)"
fi

mkdir -p "${DIST_DIR}"
find "${DIST_DIR}" -maxdepth 1 -type f -name '*.AppImage' -delete

log "Montando AppDir e dependências com linuxdeploy"

# Permite executar as próprias ferramentas AppImage mesmo em ambientes
# onde FUSE não esteja disponível.
export APPIMAGE_EXTRACT_AND_RUN=1
export ARCH="x86_64"

# O strip embutido no linuxdeploy pode ser mais antigo que o binutils
# da distribuição usada para o build. Fedora moderno usa seções ELF
# como .relr.dyn, que versões antigas de strip não reconhecem.
# O binário Go já foi compilado com -ldflags="-s -w".
export NO_STRIP=1

"${LINUXDEPLOY}" \
    --appdir "${APPDIR}" \
    --executable "${APPDIR}/usr/bin/minitela-gui" \
    --desktop-file "${APPIMAGE_DIR}/minitela.desktop" \
    --icon-file "${APPIMAGE_DIR}/minitela.svg" \
    --custom-apprun "${APPIMAGE_DIR}/AppRun"

# O teste físico no Fedora 44 demonstrou que misturar as bibliotecas
# X11 copiadas pelo linuxdeploy com libX11/libxcb/libGL do sistema
# causa SIGSEGV ainda no carregador dinâmico (_dl_init).
#
# Mantemos o RPATH preparado pelo linuxdeploy, mas removemos a família
# libX do AppDir. Sem uma cópia local, o loader usa as bibliotecas X11
# da própria distribuição.
log "Removendo bibliotecas X11 do AppDir"

find "${APPDIR}/usr/lib" \
    -maxdepth 1 \
    -type f \
    -name 'libX*.so*' \
    -print \
    -delete

# Backend e utilitário de controle são instalados somente DEPOIS do
# linuxdeploy. Assim eles não recebem RPATH/patchelf destinado ao AppDir.
# Além disso são compilados com CGO desabilitado para não depender do
# carregador dinâmico/libc da distribuição quando forem copiados para
# ~/.local/share/minitela/bin.
log "Instalando backend e MiniTelaCtl sem alterações do linuxdeploy"

install -m 0755 "${BUILD_BACKEND}" "${APPDIR}/usr/bin/minitela"
install -m 0755 "${BUILD_CTL}" "${APPDIR}/usr/bin/MiniTelaCtl"

file "${APPDIR}/usr/bin/minitela"
file "${APPDIR}/usr/bin/MiniTelaCtl"

FINAL="${DIST_DIR}/MiniTela-x86_64.AppImage"
rm -f -- "${FINAL}"

log "Gerando AppImage com appimagetool"

ARCH="x86_64" \
APPIMAGE_EXTRACT_AND_RUN=1 \
"${APPIMAGETOOL}" \
    "${APPDIR}" \
    "${FINAL}"

chmod +x "${FINAL}"

log "AppImage criado com sucesso"
ls -lh "${FINAL}"

echo
echo "SHA256:"
sha256sum "${FINAL}"

echo
echo "Teste:"
echo "  ${FINAL}"
