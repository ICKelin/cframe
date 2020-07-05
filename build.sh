go build -o dist/controller controller/*.go
cp controller/config.toml dist/controller.toml

go build -o dist/edage edage/*.go
cp edage/config.toml dist/edage.toml