# Codenames

Hosted at https://codenames-socket.herokuapp.com/

Simple implementation of the popular codenames app using
* Typescript / React / React-Router
* Golang / [Gorilla Router](https://github.com/gorilla/mux) / [Gorilla Websockets](https://github.com/gorilla/websocket)
* Redis for transient game storage

## Building and Running with Docker

```bash
$ docker build -t codenames .
$ docker run --rm -it -e PORT=8080 -e REDISCLOUD_URL=redis://rediscloud:@host.docker.internal:6379 -p 3000:8080 codenames
```

This assumes you have a Redis instance running locally on your machine exposed at port 6379. You can easily configure the above REDISCLOUD_URL variable to point to remote Redis instances as well.

## Running client and server separately

In a deployed environment, this codenames app runs as a [SPA](https://www.wikiwand.com/en/Single-page_application) with all static assets provided by the server. For development, it's convenient
the run these services separately.

Running golang server:

```bash
$ (server) REDISCLOUD_URL=redis://rediscloud:@localhost:6379 go run codenames
```

Running React frontend server:

```bash
$ (client) npm start
```

## Running tests

What kind of project do you think this was
