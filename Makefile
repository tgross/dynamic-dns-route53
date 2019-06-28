MAKEFLAGS += --warn-undefined-variables
SHELL:=/bin/bash
.DEFAULT_GOAL := help

VERSION := 0.1.0

.PHONY: help setup

## display this help message
help:
	@echo -e "\033[32m"
	@echo "Targets in this Makefile build dynamic-dns-route53"
	@echo
	@awk '/^##.*$$/,/[a-zA-Z_-]+:/' $(MAKEFILE_LIST) | awk '!(NR%2){print $$0p}{p=$$0}' | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}' | sort

## build the binary
build: dynamic-dns-route53

dynamic-dns-route53:
	go build

## package the binary for release
release: dynamic-dns-route53
	tar -czf dynamic-dns-route53-$(VERSION).tar.gz ./dynamic-dns-route53
	sha256sum ./dynamic-dns-route53-0.1.0.tar.gz > dynamic-dns-route53-$(VERSION).sha256

## remove binaries and packages
clean:
	rm -f dynamic-dns-route53
	rm -f dynamic-dns-route53.tar.gz
	rm -f dynamic-dns-route53-$(VERSION).sha256
