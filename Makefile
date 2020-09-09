ifeq ($(CI_COMMIT_REF_NAME),)
	branch 	= $(shell git rev-parse --abbrev-ref HEAD)
else
	branch 	= master(tag:$(CI_COMMIT_REF_NAME))
endif

commit 		= $(shell git log --pretty=format:'%H' -n 1)
now 		= $(shell date "+%Y-%m-%d %T UTC%z")
compiler 	= $(shell go version)


IMAGE_NAME := thetreep/toolbox


all: test build image

test:
	@echo "Running tests"
	@docker-compose -f docker-compose.test.yml up 	\
	 		--build 								\
			--abort-on-container-exit				\
			--force-recreate 						\
			--quiet-pull							\
			--no-color								\
			--remove-orphans 						\
			--timeout 20
	@docker-compose -f docker-compose.test.yml down
	@docker-compose -f docker-compose.test.yml rm -f

build:
	@echo "Compiling the binaries"
	CGO_ENABLED=0 									\
	GOBIN=$(PWD)/bin								\
	go install  -v									\
	    -ldflags                              		\
			"-X 'main.branch=$(branch)'        		\
			-X 'main.sha=$(commit)'           		\
			-X 'main.compiledAt=$(now)'       		\
			-X 'main.compiler=$(compiler)'			\
			-s -w"   								\
	    -a -installsuffix cgo ./...


image:
	@(echo "Building $(IMAGE_NAME) Docker Image")
	@(docker build -f Dockerfile -t $(IMAGE_NAME) .)

