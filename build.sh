export GOOS=linux 

go build -o dist/controller controller/*.go
go build -o dist/edge edge/*.go