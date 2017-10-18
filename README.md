## Text messaging framework

This service handles sending and receiving text messages via Twilio for the Rapid Response Network product. It is highly concurrent and batch based so that it throttles communication appropriately to the external API. Documentation lives in conflucence [here](https://credomobile.atlassian.net/wiki/display/Pol/Rapid+Responders).

This tool is written in GoLang for purposes of speed, expediency, and the ability to use built-in testing frameworks. We have unit tested all pieces of the audit framework to ensure that it does what we think it does. In order to contribute, please review the [Getting Started](https://golang.org/doc/install) documentation on GoLang, and source the gopath file to boot up a workspace.

### Dependency management

We use [Glide](https://github.com/Masterminds/glide) for dependency management of go packages, and you'll need to install it in order to be able to build the project. If you're on a Mac it's simple with HomeBrew, but there are also RPM packages for linux installs and pre-compiled binaries for Windoze. Once you check out the repo, simply run:

```
cd src && glide install
```

This will set up all required dependencies and enable you to build the project with standard go tooling or the provided build shell script.

To add another dependency, it's highly recommended that you use glide's CLI to do so. Simply ```cd src``` and run ```glide get github.com/my/dep```. Try to keep this minimal though, since you are depending on third party source code. The current packages we use are (URL's available in ```src/glide.yaml```):

- *Vestigo*: A very simple middleware router project. This guy is responsible for URL path matching. Since this component is super simple there's no need for heavyweight HTTP frameworks.

### Adding more code

The repo is structured in the "GoLang" way, and to add new code simply add more packages to the ```credomobile.com``` folder. Note that the ```server``` is the main package, so be sure to update any common runtime functionality there instead of creating another main package (GoLang will yell at you if you do).

To add another library package simply create it under the ```credomobile.com``` directory and go to town. Use the ```handlers``` package as a guide, and be sure to document what you do. If you wish to include your library in the build be sure to add it to the ```LIBS``` array in ```build.sh``` as follows:

```bash
# Add your new stuff here separated with spaces (in the bash way)
declare -a LIBS=("credomobile.com/texter/handlers" "credomobile.com/my/new/library")
```

### Testing

Everything should be tested. We write automated tests all our HTTP handlers, middleware, business logic code, etcetera. The only thing not tested in this applciation during the time of writing is the server routes--which we are planning on integration testing sometime in the future.

Keep mocks for interfaces in the same package as the interface declarations. This makes it easier if you add a method to an interface to remember to add a mock seam for that method. Examples abound in the ```twilio``` package.

Inspect the ```build.sh``` script for a taste of how we execute test cases and generate API documentation. In addition to this, there exists an ```integration_tests.sh``` file that will run end to end tests in a real docker-composed'd environment. These tests run fairly quickly (less than a minute); please use them.

### Running the application

This application is set up to run within Docker and has the capability of communicating to other pieces of the RRN system. There exists a specific network required so that the API can communicate with other pieces of the RRN system. To create this network, please run the following command:

```
docker network create rrn-net
```

To run the API, simply execute the command below. Note that you may wish to omit the ```-d``` during development. This enables you to view console spew as the application runs. 

```
# Use -d for daemonized execution
docker-compose up -d
```

We recommend always using ```docker-compose``` to execute the application. This prevents configuration drift of your local development machine and simulates a more production-like environment.

### Environment variables

All environment variables are passed into the docker containers at run time as indicated by the ```docker-compose.yml``` file. Please do not check in ```.env``` files as they may have sensitive information in them that doesn't belong in a source control repo.

### Security

The texter doesn't implement oAuth or any similar protocol for communication because we don't see the need. Instead, we have implemented a very simple "shared secret" authentication schema. This is built inside the texter with a simple middleware that encrypts outbout messages and decrypts inbound messages using [HMAC](http://www.ietf.org/rfc/rfc2104.txt).

### API documentation

Documentation is managed through the API Blueprint specification and is generated by the ```build.sh``` to HTML in the ```/docs``` folder using [aglio](https://github.com/danielgtaylor/aglio). Install it globally with:

```
npm install -g aglio
```
