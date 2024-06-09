
src := "bettertester"

default:
    just --list

tidy:
    cd {{src}} && go mod tidy

generate:
    cd {{src}} && buf generate

build:
    cd {{src}} && go build -o bin/bettertester

test:
    cd {{src}} && go test ./...

run:
    go run main.go

up *flags:
    docker-compose up --build {{flags}}

up-build *flags:
    just up --build {{flags}}

down:
    docker-compose down
