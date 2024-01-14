# DockerScript

Embed your project files into single existing Dockerfile file with respect to .dockerignore

Inspired by @adtac's [Gist](https://gist.github.com/adtac/595b5823ef73b329167b815757bbce9f)

## How to run

```
go install github.com/codenoid/DockerScript

cd your-project-which-contains-dockerfile
dockerscript -path .
# then it will generates Dockerfile.script
```
