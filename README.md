# lab
<p align="center">
	<a href="https://goreportcard.com/report/github.com/MontFerret/lab">
		<img alt="Go Report Status" src="https://goreportcard.com/badge/github.com/MontFerret/lab">
	</a>
<!-- 	<a href="https://codecov.io/gh/MontFerret/lab">
		<img alt="Code coverage" src="https://codecov.io/gh/MontFerret/lab/branch/master/graph/badge.svg" />
	</a> -->
	<a href="https://discord.gg/kzet32U">
		<img alt="Discord Chat" src="https://img.shields.io/discord/501533080880676864.svg">
	</a>
	<a href="https://github.com/MontFerret/lab/releases">
		<img alt="Lab release" src="https://img.shields.io/github/release/MontFerret/lab.svg">
	</a>
   <a href="https://microbadger.com/images/montferret/lab">
      <img alt="Dockerimages" src="https://images.microbadger.com/badges/version/montferret/lab.svg">
   </a>
	<a href="http://opensource.org/licenses/MIT">
		<img alt="MIT License" src="http://img.shields.io/badge/license-MIT-brightgreen.svg">
	</a>
</p>

``lab`` is a test runner for [Ferret](https://www.github.com/MontFerret/ferret) scripts.

![lab](https://raw.githubusercontent.com/MontFerret/lab/master/assets/landing.png)

## Features
- Parallel execution
- Support of multiple types of remote runtime (local binaries or HTTP services)
- Support of multiple types of script locations (file system, git, http)
- An arbitrary amount of HTTP endpoints for serving static files

## Quick start

```
$ docker run --mount src="$(pwd)/mytests",target=/data,type=bind montferret/lab
```

## Installation

### Binary
You can download the latest binaries from [here](https://github.com/MontFerret/lab/releases).

### Docker
```bash
$ docker pull montferret/lab:latest
```

## Usage

```bash
NAME:
   lab - run FQL test scripts

USAGE:
   lab [global options] [files...]

DESCRIPTION:
   Ferret test runner

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --files value, -f value     location of FQL script files to run [$LAB_FILES]
   --timeout value             test timeout in seconds (default: 30) [$LAB_TIMEOUT]
   --cdp value                 Chrome DevTools Protocol address (default: "http://127.0.0.1:9222") [$LAB_CDP]
   --reporter value, -r value  reporter (console, simple) (default: "console") [$LAB_REPORTER]
   --runtime value             url to remote Ferret runtime (http, https or bin) [$LAB_RUNTIME]
   --runtime-param value       params for remote Ferret runtime (--runtime-param=headers:{"KeyId": "abcd"} --runtime-param=path:"/ferret" }) [$LAB_RUNTIME_PARAM]
   --concurrency value         number of multiple tests to run at a time (default: 1) [$LAB_CONCURRENCY]
   --times value               number of times to run each test (default: 1) [$LAB_TIMES]
   --cdn value                 file or directory to serve via HTTP (./dir:8080 as default or ./dir:8080@name as named) [$LAB_CDN]
   --param value, -p value     query parameter (--param=foo:"bar", --param=id:1) [$LAB_PARAM]
   --wait value, -w value      tests and waits on the availability of remote resources (--wait http://127.0.0.1:9222/json/version --wait postgres://locahost:5432/mydb) [$LAB_WAIT]
   --wait-timeout value        wait timeout in seconds (default: 5) [$LAB_WAIT_TIMEOUT]
   --wait-attempts value       wait attempts (default: 5) [$LAB_WAIT_ATTEMPTS]
   --help, -h                  show help (default: false)
   --version, -v               print the version (default: false)
```
