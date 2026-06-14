module github.com/eddie/registry/registry/samples

go 1.25.0

// Initialized go.mod for samples

require (
	github.com/eddieraa/registry v0.4.12
	github.com/lmittmann/tint v1.1.3
	github.com/nats-io/nats.go v1.52.0
)

replace github.com/eddieraa/registry => ..

require (
	github.com/klauspost/compress v1.18.5 // indirect
	github.com/nats-io/nkeys v0.4.15 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.49.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
)
