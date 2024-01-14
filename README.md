# DockerScript

Embed your project files into single existing Dockerfile file with respect to .dockerignore & GZIP compression support

Inspired by @adtac's [Gist](https://gist.github.com/adtac/595b5823ef73b329167b815757bbce9f)

## Features

- [x] .dockerignore support
- [x] GZIP Compression Support (~80% Smaller!!!)

## How to run

Install [Go](https://dev.to/codenoid_/easiest-way-to-install-go-on-linux-gim) to follow the step below

```sh
go install github.com/codenoid/docker-script@latest # install dockerscript

cd your-project-which-contains-dockerfile

# generates Dockerfile.script which will contains all of your project files
docker-script -path .

# run the generated Dockerfile
./Dockerfile.script
```

## Is this ready for production?

yeah yeah, it's ready, but you may need to test your Dockerfile.script first, you may want to change something