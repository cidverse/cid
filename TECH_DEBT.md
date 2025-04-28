# Technical Debt

## TD-001: Missing support for SARIF in GitLab (2025-04-29)

**Context**:
GitLab currently does not support SARIF reports natively.

**Workaround**:
We call a external cli [gitlab-sarif-converter](https://gitlab.com/ignis-build/sarif-converter) to transform SARIF into a GitLab-supported format.

**Removal Plan**:
Delete the conversion step once native SARIF support is implemented in GitLab (https://gitlab.com/gitlab-org/gitlab/-/issues/452042).
