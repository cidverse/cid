# Technical Debt

## TD-001: Missing support for SARIF in GitLab (2025-04-29)

**Context**:
GitLab currently does not support SARIF reports natively.

**Workaround**:
We call a external cli [gitlab-sarif-converter](https://gitlab.com/ignis-build/sarif-converter) to transform SARIF into a GitLab-supported format.

**Removal Plan**:
Delete the conversion step once native SARIF support is implemented in GitLab (https://gitlab.com/gitlab-org/gitlab/-/issues/452042).

## TD-002: Dotnet and NPM permissions issue (2025-04-29)

**Context**:
The `dotnet` and `npm` commands don't run properly in rootless containers yet.

**Workaround**:
The `dotnet` and `npm` commands are run as root in the container.

## TD-003: Dynamic Version in Cargo.toml (2025-04-29)

**Context**:
Cargo does not allow setting dynamic versions, see https://github.com/rust-lang/cargo/issues/6583.

**Workaround**:
Manually patch the `Cargo.toml` file before running the `cargo` command. Due to this we cannot use the `--locked` flag and need `--allow-dirty`.
This might also trigger some security alerts due to source code modifications mid-build.
