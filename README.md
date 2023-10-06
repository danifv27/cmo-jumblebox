# SRE Forge - A Command Line Interface

This CLI application is a powerful and flexible tool for building robust, scalable systems that are easy to test and maintain. Its clean architecture and modular design make it highly adaptable to a wide range of use cases, and its Prometheus exporter provides valuable insights into system performance and behavior.

The application is designed to be highly modular and flexible, with a focus on testability and maintainability. It follows the principles of clean architecture, which means that it is organized into layers that are decoupled from one another and can be easily swapped out or replaced without affecting the overall functionality of the system.

At the core of the application is the business logic layer (aka usecase), which contains all the domain-specific code that drives the application's behavior. This layer is responsible for implementing the use cases and business rules of the application, and it communicates with the outside world through well-defined interfaces.

On top of the business logic layer is the application layer, which provides a high-level API for interacting with the system. This layer is responsible for translating user requests into actions that can be performed by the business logic layer, and it handles things like input validation and error handling.

The CLI component of the application is built on top of the application layer, providing a command-line interface that users can interact with to perform various actions. The CLI is designed to be simple and intuitive, with clear and concise commands that users can easily remember and use.

# Quick Start

Installation of `SREForge` is dead-simple, just download the release for your system and run the binary. The binaries are published in Artifactory, under `pc-maven/com/adidas/devops/SREForge/` package.

The binaries are classified by operating system (darwin, linux or windows), architecture (amd64) and version.

## Download the binary 

```bash
export VERSION=0.0.1
export GOOS=darwin
export GOARCH=amd64

curl -Lo SREForge https://tools.adidas-group.com/artifactory/pc-maven/com/adidas/devops/SREForge/${GOOS}/${GOARCH}/${VERSION}/bin/SREForge

# make the binary executable
chmod +x SREForge
```

## Build from source

Clone this repo and:

```bash
git clone https://bitbucket.tools.3stripes.net/scm/~daniel.fraile_adidas.com/sre-forge.git
cd cmd/<tool-name>
make local

# to 'install' the binary, make it executable and either call it directy, put 
# it in your PATH, or move it to a location which is already in your PATH:

chmod +x SREForge
mv SREForge /usr/local/bin
```

### Cross compile

Use `make build-ARCH` to cross-compile to a diferent architecture. Currently available architectures are:
 
 * linux-amd64
 * darwin-amd64
 * windows-amd64

When crosscompiling to Windows you may need to install a compiler. If you are running on Mac OS:

```bash
brew install mingw-w64
```

### Makefile rules

When developing, you can use the `Makefile` for doing the following operations:

| Name                 | Description                                                      |
| --------------------:| -----------------------------------------------------------------|
| `init`               | Initialize the module                                            |
| `clean`              | Clean out all generated items                                    |
| `clean-<ARCH>`       | Clean out all generated items                                    |
| `coverage`           | Generates the total code coverage of the project                 |
| `package`            | Build final docker image with just the go binary inside          |
| `tag`                | Tag image created by package with latest, git commit and version |
| `push`               | Push tagged images to registry                                   |
| `test`               | Run tests on a compiled project                                  |
| `unit_test`          | Run all available unit tests.                                    |
| `labels`             | Show image labels                                         |
| `build-<ARCH>`       | Build application for a specific arch.                           |
| `local`              | Build application for the local arch                             |
| `artifactory-<ARCH>` | Push binary to Artifactory                                       |


# Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://bitbucket.tools.3stripes.net/scm/~daniel.fraile_adidas.com/sre-forge.git). 

# Acknowledgement
