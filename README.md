# compose-telepresence CLI
Simple command line which wraps Telepresence to test Compose `provider` services

This CLI provide `compose up` and `compose down` commands which Docker Compose will run to:
- Setup the Telepresence configuration to the current Kubernetes context
- Create intercept for `provider` service defined in the Compose file
- Clean intercept and remove traffic manager on Compose down command

## build

```shell
make 
```
## install
Change the directory were you want to install the binary in `Makefile`
```shell
make install
```

## help
To know all the command available
```shell
make help
```