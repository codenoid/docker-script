# DockerScript

Embed your project files into single existing Dockerfile file with respect to .dockerignore

Inspired by @adtac's [Gist](https://gist.github.com/adtac/595b5823ef73b329167b815757bbce9f)

## How to run

```sh
go install github.com/codenoid/docker-script@latest # install dockerscript

cd your-project-which-contains-dockerfile

# generates Dockerfile.script which will contains all of your project files
docker-script -path .

# run the generated Dockerfile

./Dockerfile.script
```
