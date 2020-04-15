# drelayer

`drelayer` is a standalone Go daemon that allows any Handshake TLD owner to offer portions of their DDRP blob space to other people.


## Installation From Source

1. Install the `go` toolchain.
2. Install `postgres`.
3. Install the required dependencies:
	- `go-swagger`.
	- `migrate`.
4. Create a `postgres` database to contain `drelayer`'s data.
5. From the root of the project, run  `migrate -source file://./store/migrations -database <your-databse-dsn> -verbose up`.
6. Run `make drelayer` to build the daemon.
7. Start a DDRP node.
8. Copy `config.example.yml` to `config.yml`, and modify the values as appropriate for your system.
9. Start `drelayer` by running `./build/drelayer -c <config-path>`.

## REST API

`drelayer` exposes a REST API. You can generate API stubs and documentation from the `swagger.yml` file in the `swagger` directory, or paste its contents into the online [Swagger Editor](https://editor.swagger.io).

## Useful Things

To extract `ddrpcli` private keys for use with the relayer, run `xxd -c 64 -p ~/.ddrpcli/identity`.