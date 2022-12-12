# Castopod Operator

## How to contribute
### Create a local environment
```SHELL
kind create cluster --config ./kind.yaml
```
### Deploy CRD and Role
```SHELL
task local:install
task local:deploy
```