# Continuous Integration and Deployment Workflow CLI - `cid`

[![Go Reference](https://pkg.go.dev/badge/github.com/cidverse/cid.svg)](https://pkg.go.dev/github.com/cidverse/cid)
[![Go Report Card](https://goreportcard.com/badge/github.com/cidverse/cid)](https://goreportcard.com/report/github.com/cidverse/cid)

Run your continuous integration and deployment workflows in a platform-agnostic way!

- **Platform Agnostic** - Your workflow works locally and on any repository/pipeline service of your choice. -> the [normalize.ci](https://github.com/cidverse/normalizeci) component normalizes all environment variables into a global format.
- **Fast Feedback** - Rather than having to commit/push/wait every time you want to test out the changes you are making to your continuous integration and deployment process, you can use `cid` to run/test your workflow locally. `cid` can provide normalized environment variables as the ci service would based on scm repository information.

# How Does It Work?

When you run `cid` it searches for the repository based on your working directory, by looking for repository folders (ie. `.git`) in each parent directory of your current working directory.

The workflow `actions` can detect the current project type and build your project if they follow some simple conventions, if you have more complex projects that contain many different modules you can configure a custom workflow in the `cid.yml`'s workflow section.

# Installation

**WIP**

# Stages

## Main Workflow

| Stage       |                                       Description                                        |
|-------------|:----------------------------------------------------------------------------------------:|
| build       |                                    builds the project                                    |
| test        |                                      runs the tests                                      |
| sast        |             analyzes source code to flag programming errors and other issues             |
| package     |                         packages the artifacts for distribution                          |
| qualitygate | a final quality gate to check for any kind of security / quality issues before releasing |
| publish     |                       publishes the previously packaged artifacts                        |
| deploy      |           runs a deployment onto the runtime environment (microservices, ...)            |

## Virtual (Helper)

| Stage |                            Description                            |
|-------|:-----------------------------------------------------------------:|
| init  | helps with the general project setup (git hooks, gitignore, etc.) |

# Configuration

You can place a `cid.yml` in your project root directory (-> scm repository root directory).
