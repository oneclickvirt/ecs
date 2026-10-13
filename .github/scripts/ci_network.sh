# shellcheck shell=bash
# CI network operations have bounded retries; build and test commands are not retried.
ci_retry() {
  local attempt status stdout_file stderr_file
  stdout_file=$(mktemp)
  stderr_file=$(mktemp)
  for attempt in 1 2 3; do
    if "$@" >"$stdout_file" 2>"$stderr_file"; then
      cat "$stdout_file"
      cat "$stderr_file" >&2
      rm -f "$stdout_file" "$stderr_file"
      return 0
    else
      status=$?
    fi
    if [[ "$attempt" == 3 ]] || ! grep -Eiq 'HTTP[^0-9]*(408|429|500|502|503|504)|timed? out|timeout|connection (reset|refused|closed)|temporary failure|TLS handshake|unexpected EOF|could not resolve host|failed to connect|remote end hung up' "$stdout_file" "$stderr_file"; then
      cat "$stdout_file"
      cat "$stderr_file" >&2
      rm -f "$stdout_file" "$stderr_file"
      return "$status"
    fi
    sleep "$((attempt * 2))"
  done
}
gh() { ci_retry command gh "$@"; }
git() {
  if [[ "${1:-}" != clone ]]; then
    command git "$@"
    return
  fi
  local destination temporary status
  local clone_args=("$@")
  destination="${clone_args[${#clone_args[@]}-1]}"
  # Calls without an explicit destination keep git's native behavior.
  if [[ "$destination" == *://* || "$destination" == *.git ]]; then
    ci_retry command git "$@"
    return
  fi
  if [[ -e "$destination" || -L "$destination" ]]; then
    printf 'Clone destination already exists.\n' >&2
    return 1
  fi
  mkdir -p "$(dirname "$destination")"
  temporary=$(mktemp -d "${destination}.ci-XXXXXX")
  clone_args[${#clone_args[@]}-1]="$temporary"
  # A failed clone may leave a partial checkout; clear only our owned temp directory.
  if ci_retry ci_clone_attempt "$temporary" "${clone_args[@]}"; then
    if [[ -e "$destination" || -L "$destination" ]]; then
      rm -rf "$temporary"
      return 1
    fi
    mv "$temporary" "$destination"
  else
    status=$?
    rm -rf "$temporary"
    return "$status"
  fi
}
ci_clone_attempt() {
  local temporary="$1"
  shift
  rm -rf "$temporary"
  command git "$@"
}
curl() {
  local argument mutating=false status stdout_file stderr_file source_index=-1 backup='' i
  local curl_args=("$@")
  for ((i=0; i<${#curl_args[@]}; i++)); do
    argument="${curl_args[i]}"
    case "$argument" in
      --data*|-d*|--form*|-F*|--upload-file*|-T*) mutating=true ;;
      POST|PATCH|PUT|DELETE|-XPOST|-XPATCH|-XPUT|-XDELETE|--request=POST|--request=PATCH|--request=PUT|--request=DELETE) mutating=true ;;
      https://*) source_index=$i ;;
    esac
  done
  if $mutating; then
    # A timed-out mutation may already have reached the server.
    command curl --fail-with-body --connect-timeout 15 --max-time 600 "${curl_args[@]}"
    return
  fi
  if [[ "$source_index" -ge 0 ]]; then
    case "${curl_args[source_index]}" in
      https://cdn.spiritlhl.net/https://*|https://cdn.spiritlhl.top/https://*)
        backup="${curl_args[source_index]#*/https://}"
        backup="https://$backup" ;;
      https://ftp.gnu.org/gnu/*)
        backup="https://ftpmirror.gnu.org/${curl_args[source_index]#https://ftp.gnu.org/gnu/}" ;;
    esac
  fi
  stdout_file=$(mktemp)
  stderr_file=$(mktemp)
  if command curl --fail-with-body --connect-timeout 15 --max-time 1200 --retry 3 --retry-max-time 180 --retry-connrefused "${curl_args[@]}" >"$stdout_file" 2>"$stderr_file"; then
    cat "$stdout_file"
    cat "$stderr_file" >&2
    rm -f "$stdout_file" "$stderr_file"
    return 0
  else
    status=$?
  fi
  if [[ -n "$backup" ]]; then
    curl_args[source_index]="$backup"
    if command curl --fail-with-body --connect-timeout 15 --max-time 1200 --retry 3 --retry-max-time 180 --retry-connrefused "${curl_args[@]}" >"$stdout_file" 2>"$stderr_file"; then
      cat "$stdout_file"
      cat "$stderr_file" >&2
      rm -f "$stdout_file" "$stderr_file"
      return 0
    else
      status=$?
    fi
  fi
  cat "$stdout_file"
  cat "$stderr_file" >&2
  rm -f "$stdout_file" "$stderr_file"
  return "$status"
}
wget() { command wget --tries=3 --timeout=45 --waitretry=2 --retry-connrefused "$@"; }
