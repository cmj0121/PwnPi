include Makefile.in

.PHONY: all clean test run build upgrade help

all: 			# default action
	@pre-commit install --install-hooks
	@git config commit.template .git-commit-template

clean:			# clean-up environment
	@find . -name '*.sw[po]' -delete
	@rm -f pwnpi

test:			# run test
	go mod tidy
	go test -v ./...

run:			# run in the local environment
	go mod tidy
	go run cmd/pwnpi/main.go

build:			# build the binary/library
	go build -ldflags "-s -w" -o pwnpi cmd/pwnpi/main.go

upgrade:		# upgrade all the necessary packages
	pre-commit autoupdate

help:			# show this message
	@printf "Usage: make [OPTION]\n"
	@printf "\n"
	@perl -nle 'print $$& if m{^[\w-]+:.*?#.*$$}' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?#"} {printf "    %-18s %s\n", $$1, $$2}'

.PHONY: prologue install
SSH_PUBLIC_KEY := ~/.ssh/id_ed25519.pub

prologue:		# setup everything before access your PwnPi
	@ssh-copy-id -i $(SSH_PUBLIC_KEY) $(USERNAME)@$(HOSTNAME)

install:		# sync and instal the package to your PwnPi
	env GOOS=linux GOARCH=arm GOARM=7 go build -ldflags "-s -w" -o pwnpi cmd/pwnpi/main.go
	rsync -az pwnpi $(USERNAME)@$(HOSTNAME):~
