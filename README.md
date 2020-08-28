# L2c-Operator
![l2c-operator](https://github.com/tmax-cloud/l2c-operator/workflows/l2c-operator/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/tmax-cloud/l2c-operator)](https://goreportcard.com/report/github.com/tmax-cloud/l2c-operator)
![release](https://img.shields.io/github/v/release/tmax-cloud/l2c-operator)

## Legacy-to-Cloud (L2c)
Legacy-to-Cloud is a module to migrate applications running in a legacy environment to a Kubernetes environment.

## Start using L2c-operator
- [Installation Guide](./docs/installation.md)
- [Quick Start Guide](./docs/quickstart.md)

## Supported Features
### WAS Migration Analyze (T-up Jeus)
- Now supports Weblogic &rightarrow; Jeus
- Provides incompatibilities of existing codes on the target WAS

### DB Migration (T-up Tibero)
- Now supports Oracle &rightarrow; Tibero
- Deploys a DB deployment and migrates data from source to target

### Web IDE (VS Code)
- If WAS migration analysis reports issues, Web IDE is automatically deployed. The IDE employs SonarLint. 

### Build/Deploy
- Build the source using S2I and deploy it to the cluster.
