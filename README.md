# DockerScript

Embed your project files into single existing Dockerfile file with respect to .dockerignore

## How to run

```
go install github.com/codenoid/DockerScript

cd your-project-which-contains-dockerfile
dockerscript -path .
# then it will generates Dockerfile.script
```
