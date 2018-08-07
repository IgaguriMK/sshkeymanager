.PHONY: all
all: build


.PHONY: build
build:
	go build cmd/sshkeymanager/sshkeymanager.go
