#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2026 Rillan AI LLC
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

# ──────────────────────────────────────────────────────────────────────────────
# LLM Skills Installer
#
# Detects installed LLM coding tools (Claude Code, Codex, OpenCode) and
# installs skills at the user level for each detected tool.
#
# Usage:
#   ./install.sh                          # Auto-detect tools, install all skills
#   ./install.sh --list                   # List detected tools and available skills
#   ./install.sh --uninstall              # Remove installed skills from all tools
#   ./install.sh --tool=claude            # Install only for Claude Code
#   ./install.sh --tool=claude,opencode   # Install for Claude Code and OpenCode
#   ./install.sh --lang=go               # Install only Go skills
#   ./install.sh --lang=go,python         # Install Go and Python skills
#   ./install.sh --dry-run               # Show what would be installed without doing it
# ──────────────────────────────────────────────────────────────────────────────

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors (disabled if not a terminal)
if [ -t 1 ]; then
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    RED='\033[0;31m'
    BLUE='\033[0;34m'
    CYAN='\033[0;36m'
    BOLD='\033[1m'
    DIM='\033[2m'
    NC='\033[0m'
else
    GREEN='' YELLOW='' RED='' BLUE='' CYAN='' BOLD='' DIM='' NC=''
fi

log_info()    { echo -e "${GREEN}[+]${NC} $*"; }
log_warn()    { echo -e "${YELLOW}[!]${NC} $*"; }
log_error()   { echo -e "${RED}[-]${NC} $*"; }
log_step()    { echo -e "${BLUE}[>]${NC} $*"; }
log_header()  { echo -e "\n${BOLD}$*${NC}"; }
log_replace() { echo -e "${CYAN}[~]${NC} $*"; }
log_dry()     { echo -e "${DIM}[dry-run]${NC} $*"; }

DRY_RUN=false

# ──────────────────────────────────────────────────────────────────────────────
# Config: user-level paths for each tool
# ──────────────────────────────────────────────────────────────────────────────

CLAUDE_COMMANDS_DIR="${HOME}/.claude/commands"
CODEX_SKILLS_DIR="${HOME}/.agents/skills"
OPENCODE_AGENTS_DIR="${XDG_CONFIG_HOME:-${HOME}/.config}/opencode/agents"

VALID_TOOLS="claude codex opencode"
LANGUAGES=(go rust python terraform helm kubernetes operator adr planning docker security cicd rfc)

# ──────────────────────────────────────────────────────────────────────────────
# Version helpers
# ──────────────────────────────────────────────────────────────────────────────

# Extract version from a skill file (reads the <!-- version: X.Y.Z --> comment)
extract_version() {
    local file="$1"
    if [ -f "$file" ]; then
        sed -n 's/.*<!-- version: \([0-9][0-9]*\.[0-9][0-9]*\.[0-9][0-9]*\).*/\1/p' "$file" | head -1 || echo "unknown"
    else
        echo "none"
    fi
}

# Compare two semver strings. Returns: "newer", "same", "older", or "unknown"
compare_versions() {
    local src_ver="$1" dest_ver="$2"

    if [ "$src_ver" = "unknown" ] || [ "$dest_ver" = "unknown" ] || [ "$dest_ver" = "none" ]; then
        echo "unknown"
        return
    fi

    if [ "$src_ver" = "$dest_ver" ]; then
        echo "same"
        return
    fi

    # Split into components
    local IFS='.'
    read -r s_major s_minor s_patch <<< "$src_ver"
    read -r d_major d_minor d_patch <<< "$dest_ver"

    if [ "$s_major" -gt "$d_major" ] 2>/dev/null; then echo "newer"; return; fi
    if [ "$s_major" -lt "$d_major" ] 2>/dev/null; then echo "older"; return; fi
    if [ "$s_minor" -gt "$d_minor" ] 2>/dev/null; then echo "newer"; return; fi
    if [ "$s_minor" -lt "$d_minor" ] 2>/dev/null; then echo "older"; return; fi
    if [ "$s_patch" -gt "$d_patch" ] 2>/dev/null; then echo "newer"; return; fi
    if [ "$s_patch" -lt "$d_patch" ] 2>/dev/null; then echo "older"; return; fi
    echo "same"
}

# ──────────────────────────────────────────────────────────────────────────────
# Detection
# ──────────────────────────────────────────────────────────────────────────────

detect_claude() {
    command -v claude &>/dev/null || [ -d "${HOME}/.claude" ]
}

detect_codex() {
    command -v codex &>/dev/null || [ -d "${HOME}/.codex" ]
}

detect_opencode() {
    command -v opencode &>/dev/null || [ -d "${XDG_CONFIG_HOME:-${HOME}/.config}/opencode" ]
}

get_detected_tools() {
    local tools=()
    detect_claude   && tools+=(claude)
    detect_codex    && tools+=(codex)
    detect_opencode && tools+=(opencode)
    echo "${tools[*]}"
}

# ──────────────────────────────────────────────────────────────────────────────
# Skill name helpers
# ──────────────────────────────────────────────────────────────────────────────

skill_basename() { echo "${1%.skill.md}"; }
skill_name()     { echo "${1}-$(skill_basename "$2")"; }

list_skill_files() {
    local lang="$1"
    local dir path files=()
    dir="${SCRIPT_DIR}/${lang}"

    for path in "${dir}"/*.skill.md; do
        [ -e "$path" ] || continue
        files+=("${path##*/}")
    done

    if [ "${#files[@]}" -eq 0 ]; then
        return 0
    fi

    printf '%s\n' "${files[@]}" | sort
}

# Get the title from a skill file (first heading, skipping version comment)
skill_title() {
    local file="$1"
    grep -m1 '^# ' "$file" 2>/dev/null | sed 's/^# //'
}

# Get the destination path for a skill given the tool
skill_dest() {
    local tool="$1" lang="$2" file="$3"
    local name
    name="$(skill_name "$lang" "$file")"
    case "$tool" in
        claude)   echo "${CLAUDE_COMMANDS_DIR}/${name}.md" ;;
        codex)    echo "${CODEX_SKILLS_DIR}/${name}/SKILL.md" ;;
        opencode) echo "${OPENCODE_AGENTS_DIR}/${name}.md" ;;
    esac
}

# ──────────────────────────────────────────────────────────────────────────────
# Replacement check — warns user and shows version diff
# ──────────────────────────────────────────────────────────────────────────────

check_replacement() {
    local tool="$1" lang="$2" file="$3" src="$4"
    local dest name src_ver dest_ver cmp

    dest="$(skill_dest "$tool" "$lang" "$file")"
    name="$(skill_name "$lang" "$file")"

    if [ ! -f "$dest" ]; then
        return 1  # No existing file, not a replacement
    fi

    src_ver="$(extract_version "$src")"
    dest_ver="$(extract_version "$dest")"
    cmp="$(compare_versions "$src_ver" "$dest_ver")"

    case "$cmp" in
        same)
            log_replace "${name}: already at v${dest_ver} ${DIM}(no change)${NC}"
            return 0
            ;;
        newer)
            log_replace "${name}: upgrading v${dest_ver} -> v${src_ver}"
            return 0
            ;;
        older)
            log_warn "${name}: installed v${dest_ver} is newer than source v${src_ver} ${DIM}(downgrade)${NC}"
            return 0
            ;;
        *)
            log_replace "${name}: replacing existing ${DIM}(version unknown)${NC}"
            return 0
            ;;
    esac
}

# ──────────────────────────────────────────────────────────────────────────────
# Install functions
# ──────────────────────────────────────────────────────────────────────────────

install_claude_skill() {
    local lang="$1" file="$2"
    local name src dest
    name="$(skill_name "$lang" "$file")"
    src="${SCRIPT_DIR}/${lang}/${file}"
    dest="${CLAUDE_COMMANDS_DIR}/${name}.md"

    check_replacement "claude" "$lang" "$file" "$src" || true

    if $DRY_RUN; then
        log_dry "Would install Claude Code command ${BOLD}/${name}${NC} -> ${dest}"
        return
    fi

    mkdir -p "${CLAUDE_COMMANDS_DIR}"
    cp "$src" "$dest"
    log_info "Claude Code: installed ${BOLD}/${name}${NC} command"
}

install_codex_skill() {
    local lang="$1" file="$2"
    local name src dest_dir dest title
    name="$(skill_name "$lang" "$file")"
    src="${SCRIPT_DIR}/${lang}/${file}"
    dest_dir="${CODEX_SKILLS_DIR}/${name}"
    dest="${dest_dir}/SKILL.md"

    check_replacement "codex" "$lang" "$file" "$src" || true

    if $DRY_RUN; then
        log_dry "Would install Codex skill ${BOLD}\$${name}${NC} -> ${dest}"
        return
    fi

    mkdir -p "$dest_dir"
    title="$(skill_title "$src")"

    # Codex requires SKILL.md with name + description frontmatter.
    # Only treat the file as having frontmatter if `---` is the very first line —
    # source files often contain `---` YAML separators inside example code blocks.
    if [ "$(head -1 "$src")" = "---" ]; then
        cp "$src" "$dest"
    else
        {
            echo "---"
            echo "name: ${name}"
            echo "description: ${title}"
            echo "---"
            echo ""
            cat "$src"
        } > "$dest"
    fi

    log_info "Codex: installed ${BOLD}\$${name}${NC} skill"
}

install_opencode_skill() {
    local lang="$1" file="$2"
    local name src dest title description body
    name="$(skill_name "$lang" "$file")"
    src="${SCRIPT_DIR}/${lang}/${file}"
    dest="${OPENCODE_AGENTS_DIR}/${name}.md"

    check_replacement "opencode" "$lang" "$file" "$src" || true

    if $DRY_RUN; then
        log_dry "Would install OpenCode agent ${BOLD}@${name}${NC} -> ${dest}"
        return
    fi

    mkdir -p "${OPENCODE_AGENTS_DIR}"

    if [ "$(head -1 "$src")" = "---" ]; then
        # Source has frontmatter — extract its description and emit body without source frontmatter.
        # OpenCode wants its own `description` (quoted); `name:` from source is dropped.
        description="$(awk '
            BEGIN { n = 0 }
            /^---$/ { n++; next }
            n == 1 && /^description:[[:space:]]*/ {
                sub(/^description:[[:space:]]*/, "")
                gsub(/^"|"$/, "")
                print
                exit
            }
        ' "$src")"
        body="$(awk 'BEGIN { n = 0 } /^---$/ { n++; next } n >= 2 { print }' "$src")"
        {
            echo "---"
            echo "description: \"${description}\""
            echo "---"
            echo ""
            echo "$body"
        } > "$dest"
    else
        title="$(skill_title "$src")"
        {
            echo "---"
            echo "description: \"${title}\""
            echo "---"
            echo ""
            cat "$src"
        } > "$dest"
    fi

    log_info "OpenCode: installed ${BOLD}@${name}${NC} agent"
}

# ──────────────────────────────────────────────────────────────────────────────
# Uninstall functions
# ──────────────────────────────────────────────────────────────────────────────

uninstall_claude_skill() {
    local lang="$1" file="$2"
    local name dest
    name="$(skill_name "$lang" "$file")"
    dest="${CLAUDE_COMMANDS_DIR}/${name}.md"
    if [ -f "$dest" ]; then
        if $DRY_RUN; then
            log_dry "Would remove Claude Code /${name}"
        else
            rm "$dest"
            log_info "Claude Code: removed /${name}"
        fi
    fi
}

uninstall_codex_skill() {
    local lang="$1" file="$2"
    local name dest_dir
    name="$(skill_name "$lang" "$file")"
    dest_dir="${CODEX_SKILLS_DIR}/${name}"
    if [ -d "$dest_dir" ]; then
        if $DRY_RUN; then
            log_dry "Would remove Codex \$${name}"
        else
            rm -rf "$dest_dir"
            log_info "Codex: removed \$${name}"
        fi
    fi
}

uninstall_opencode_skill() {
    local lang="$1" file="$2"
    local name dest
    name="$(skill_name "$lang" "$file")"
    dest="${OPENCODE_AGENTS_DIR}/${name}.md"
    if [ -f "$dest" ]; then
        if $DRY_RUN; then
            log_dry "Would remove OpenCode @${name}"
        else
            rm "$dest"
            log_info "OpenCode: removed @${name}"
        fi
    fi
}

# ──────────────────────────────────────────────────────────────────────────────
# Main operations
# ──────────────────────────────────────────────────────────────────────────────

do_list() {
    log_header "Detected LLM Coding Tools"
    echo ""
    local found=0
    if detect_claude; then
        log_info "Claude Code  $(command -v claude 2>/dev/null || echo "(config dir found)")"
        found=1
    fi
    if detect_codex; then
        log_info "Codex        $(command -v codex 2>/dev/null || echo "(config dir found)")"
        found=1
    fi
    if detect_opencode; then
        log_info "OpenCode     $(command -v opencode 2>/dev/null || echo "(config dir found)")"
        found=1
    fi
    if [ "$found" -eq 0 ]; then
        log_warn "No supported tools detected"
    fi

    echo ""
    log_header "Available Skills"
    echo ""
    printf "  ${BOLD}%-22s %-8s %s${NC}\n" "SKILL" "VERSION" "TITLE"
    printf "  %-22s %-8s %s\n" "─────" "───────" "─────"
    for lang in "${LANGUAGES[@]}"; do
        while IFS= read -r file; do
            local name ver title
            name="$(skill_name "$lang" "$file")"
            ver="$(extract_version "${SCRIPT_DIR}/${lang}/${file}")"
            title="$(skill_title "${SCRIPT_DIR}/${lang}/${file}")"
            printf "  %-22s %-8s %s\n" "$name" "v${ver}" "$title"
        done < <(list_skill_files "$lang")
    done
}

do_install() {
    local filter_tools_raw="${1:-}"
    local filter_langs_raw="${2:-}"

    # Parse comma-separated tool list
    local tools
    if [ -n "$filter_tools_raw" ]; then
        tools="${filter_tools_raw//,/ }"
        # Validate each tool
        for t in $tools; do
            if ! echo "$VALID_TOOLS" | grep -qw "$t"; then
                log_error "Unknown tool: ${t}"
                echo "Available tools: ${VALID_TOOLS}"
                exit 1
            fi
        done
    else
        tools="$(get_detected_tools)"
    fi

    if [ -z "$tools" ]; then
        log_error "No supported LLM tools detected on this system."
        echo ""
        echo "Supported tools:"
        echo "  - Claude Code  (https://github.com/anthropics/claude-code)"
        echo "  - Codex        (https://github.com/openai/codex)"
        echo "  - OpenCode     (https://github.com/opencode-ai/opencode)"
        echo ""
        echo "Install one of these tools first, or use --tool=<name> to force install."
        exit 1
    fi

    # Parse comma-separated language list
    local langs_to_install=("${LANGUAGES[@]}")
    if [ -n "$filter_langs_raw" ]; then
        IFS=',' read -ra langs_to_install <<< "$filter_langs_raw"
        for l in "${langs_to_install[@]}"; do
            if ! printf '%s\n' "${LANGUAGES[@]}" | grep -qx "$l"; then
                log_error "Unknown language: ${l}"
                echo "Available: ${LANGUAGES[*]}"
                exit 1
            fi
        done
    fi

    if $DRY_RUN; then
        log_header "Dry Run — Nothing Will Be Modified"
    else
        log_header "Installing Skills"
    fi
    echo ""

    local installed_count=0
    for tool in $tools; do
        log_step "Installing for ${BOLD}${tool}${NC}"
        for lang in "${langs_to_install[@]}"; do
            while IFS= read -r file; do
                local src="${SCRIPT_DIR}/${lang}/${file}"
                if [ ! -f "$src" ]; then
                    log_warn "Skill file not found: ${src} (skipping)"
                    continue
                fi
                case "$tool" in
                    claude)   install_claude_skill   "$lang" "$file" ;;
                    codex)    install_codex_skill    "$lang" "$file" ;;
                    opencode) install_opencode_skill "$lang" "$file" ;;
                    *)        log_warn "Unknown tool: $tool" ;;
                esac
                installed_count=$((installed_count + 1))
            done < <(list_skill_files "$lang")
        done
        echo ""
    done

    if $DRY_RUN; then
        log_header "Dry run complete — ${installed_count} skills would be installed"
        return
    fi

    log_header "Installation Complete (${installed_count} skills)"
    echo ""
    echo "Usage:"
    for tool in $tools; do
        local examples=""
        for lang in "${langs_to_install[@]}"; do
            while IFS= read -r file; do
                local stype
                stype="$(skill_basename "$file")"
                case "$tool" in
                    claude)   examples+=" /${lang}-${stype}" ;;
                    codex)    examples+=" \$${lang}-${stype}" ;;
                    opencode) examples+=" @${lang}-${stype}" ;;
                esac
            done < <(list_skill_files "$lang")
        done
        case "$tool" in
            claude)   echo -e "  ${BOLD}Claude Code${NC}: ${examples}" ;;
            codex)    echo -e "  ${BOLD}Codex${NC}:       ${examples}" ;;
            opencode) echo -e "  ${BOLD}OpenCode${NC}:    ${examples}" ;;
        esac
    done
}

do_uninstall() {
    local filter_tools_raw="${1:-}"
    local filter_langs_raw="${2:-}"

    local tools
    if [ -n "$filter_tools_raw" ]; then
        tools="${filter_tools_raw//,/ }"
    else
        # Uninstall from all tools regardless of detection
        tools="$VALID_TOOLS"
    fi

    local langs_to_uninstall=("${LANGUAGES[@]}")
    if [ -n "$filter_langs_raw" ]; then
        IFS=',' read -ra langs_to_uninstall <<< "$filter_langs_raw"
    fi

    if $DRY_RUN; then
        log_header "Dry Run — Uninstall Preview"
    else
        log_header "Uninstalling Skills"
    fi
    echo ""

    for tool in $tools; do
        for lang in "${langs_to_uninstall[@]}"; do
            while IFS= read -r file; do
                case "$tool" in
                    claude)   uninstall_claude_skill   "$lang" "$file" ;;
                    codex)    uninstall_codex_skill    "$lang" "$file" ;;
                    opencode) uninstall_opencode_skill "$lang" "$file" ;;
                esac
            done < <(list_skill_files "$lang")
        done
    done

    if $DRY_RUN; then
        echo ""
        log_info "Dry run complete"
    else
        echo ""
        log_info "Uninstall complete"
    fi
}

# ──────────────────────────────────────────────────────────────────────────────
# CLI argument parsing
# ──────────────────────────────────────────────────────────────────────────────

ACTION="install"
FILTER_TOOL=""
FILTER_LANG=""

for arg in "$@"; do
    case "$arg" in
        --list)
            ACTION="list"
            ;;
        --uninstall)
            ACTION="uninstall"
            ;;
        --dry-run)
            DRY_RUN=true
            ;;
        --tool=*)
            FILTER_TOOL="${arg#--tool=}"
            ;;
        --lang=*)
            FILTER_LANG="${arg#--lang=}"
            ;;
        --help|-h)
            cat <<'HELPEOF'
LLM Skills Installer

Installs coding skills at the user level for Claude Code, Codex, and OpenCode.

Usage: install.sh [OPTIONS]

Options:
  --list                  List detected tools and available skills
  --uninstall             Remove installed skills
  --dry-run               Show what would happen without making changes
  --tool=<t1[,t2,...]>    Install for specific tool(s): claude, codex, opencode
  --lang=<l1[,l2,...]>    Install specific skill group(s): go, rust, python, terraform, helm, kubernetes, operator, adr, planning, docker, security, cicd, rfc
  -h, --help              Show this help

Examples:
  ./install.sh                              Auto-detect tools, install all
  ./install.sh --tool=claude                Claude Code only
  ./install.sh --tool=claude,opencode       Claude Code + OpenCode
  ./install.sh --lang=go,python             Go and Python skills only
  ./install.sh --tool=codex --lang=terraform Codex + Terraform only
  ./install.sh --tool=claude --lang=helm,kubernetes Kubernetes packaging skills only
  ./install.sh --tool=codex --lang=operator  Kubernetes operator skills only
  ./install.sh --dry-run                    Preview without installing
  ./install.sh --uninstall --tool=codex     Remove skills from Codex only
HELPEOF
            exit 0
            ;;
        *)
            log_error "Unknown argument: $arg"
            echo "Use --help for usage information."
            exit 1
            ;;
    esac
done

case "$ACTION" in
    list)      do_list ;;
    install)   do_install "$FILTER_TOOL" "$FILTER_LANG" ;;
    uninstall) do_uninstall "$FILTER_TOOL" "$FILTER_LANG" ;;
esac
