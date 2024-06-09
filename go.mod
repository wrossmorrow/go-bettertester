module main

go 1.22.3

replace github.com/wrossmorrow/bettertester v0.0.0 => ./bettertester

require (
	github.com/wrossmorrow/bettertester v0.0.0 // direct
	google.golang.org/protobuf v1.34.1 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)
