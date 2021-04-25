# Tools

## Discovery

How does `CID` decide which binaries should be called / which containers hould be started to run certain commands?

There are two operation modes, suited for different cases:

## Mode: Prefer Local

The most common detection mode is searching for version specific environment variables, that are usually also set on ci workers with preinstalled software.

This also works for GitHub Actions: https://github.com/actions/virtual-environments/blob/main/images/linux/Ubuntu2004-README.md

### Golang

| Tool          | Environment   | Version  |
| ------------- |:-------------:| -----:|
| go            | GOROOT_1_16   | 1.16.0 |
| go            | GOROOT_1_15   | 1.15.0 |
| go            | GOROOT_1_14   | 1.14.0 |
| go            | GOROOT_1_13   | 1.13.0 |
| go            | GOROOT_1_12   | 1.12.0 |
| go            | GOROOT_1_11   | 1.11.0 |
| go            | GOROOT_1_10   | 1.10.0 |

### Java

| Tool          | Environment   | Version  |
| ------------- |:-------------:| -----:|
| java          | JAVA_HOME_17  | 17.0.0 |
| java          | JAVA_HOME_16  | 16.0.0 |
| java          | JAVA_HOME_15  | 15.0.0 |
| java          | JAVA_HOME_14  | 14.0.0 |
| java          | JAVA_HOME_13  | 13.0.0 |
| java          | JAVA_HOME_12  | 12.0.0 |
| java          | JAVA_HOME_11  | 11.0.0 |
| java          | JAVA_HOME_10  | 10.0.0 |
| java          | JAVA_HOME_9   | 9.0.0 |
| java          | JAVA_HOME_8   | 8.0.0 |
