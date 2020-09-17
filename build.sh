export GOOS=linux 

go build -o dist/apiserver apiserver/*.go
go build -o dist/controller controller/*.go
cp controller/config.toml dist/controller.toml

go build -o dist/edge edge/*.go
cp edge/config.toml dist/edge.toml