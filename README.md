![Repository Top Language](https://img.shields.io/github/languages/top/JavaHutt/ws-proxy)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/JavaHutt/ws-proxy)
![Github Repository Size](https://img.shields.io/github/repo-size/JavaHutt/ws-proxy)
![License](https://img.shields.io/badge/license-MIT-green)
![GitHub last commit](https://img.shields.io/github/last-commit/JavaHutt/ws-proxy)
![Coding all night)](https://img.shields.io/badge/coding-all%20night%20-purple)

<img align="right" width="30%" src="./images/tired-gopher.png">

# Proxy server component

## Task description

Task description is in [DESCRIPTION.md](DESCRIPTION.md)

## Issues found in task description and fixed

- `signal.Notify` wasn't cathing SIGTERM signal
- `proxy.OrderRequest` had a `uint8(rand.Uint32())` randomizer, which will return 1 or 2 *very* rarely. Not good for testing purposes

## Solution notes

- :trophy: standard Go library (except for Gorilla Websocket package)
- :arrow_right_hook: clean architecture (handler->service)
- :book: standard Go project layout
- :hammer: Makefile included
- :toilet: tests with mocks included

## HOWTO

- start server with 
```bash
make server
```
- then start proxy component with some restrictions 
```bash
make proxy N=5 S=7000
```
where N is a limit of opened orders per client per instrument at the moment of time
and S is the sum limit of volumes of opened orders per client per instrument at the moment of time
- finally, start the client:
```bash
make client
```
- test with
```bash
make test
```

## A picture is worth a thousand words

<img src="./images/working-example.png">
