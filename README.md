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
	<a href="https://opensource.org/licenses/Apache-2.0">
		<img alt="Apache-2.0 License" src="http://img.shields.io/badge/license-Apache-brightgreen.svg">
	</a>
</p>

``lab`` is a test runner for [Ferret](https://www.github.com/MontFerret/ferret) scripts.

Read the introductory blog post about Lab [here!](https://www.montferret.dev/blog/say-hello-to-lab/)

<p align="center">
<img alt="lab" src="https://raw.githubusercontent.com/MontFerret/lab/master/assets/landing.png" style="margin-left: auto; margin-right: auto;" width="495px" height="501px" />
</p>

## Features
- Parallel execution
- Support of multiple types of remote runtime (local binaries or HTTP services)
- Support of multiple types of script locations (file system, git, http)
- An arbitrary amount of HTTP endpoints for serving static files

## Installation

### Binary
You can download the latest binaries from [here](https://github.com/MontFerret/lab/releases).

### Shell
```bash
curl https://raw.githubusercontent.com/MontFerret/lab/master/install.sh | sh
```

### Docker
```bash
$ docker pull montferret/lab:latest
```

## Quick start

The easiest way to use ``lab`` is to execute FQL scripts as is:

```bash
$ lab myscript.fql
```

You can also pass a path to a folder that contains ``.fql`` scripts:

```bash
$ lab myscripts/
```

### Test suites

``lab`` also allows you to define suite tests in YAML:

```yaml
query:
  text: |
    LET doc = DOCUMENT("https://github.com/", { driver: "cdp" })
    
    HOVER(doc, ".HeaderMenu-details")
    CLICK(doc, ".HeaderMenu a")
    
    WAIT_NAVIGATION(doc)
    WAIT_ELEMENT(doc, '.IconNav')
    
    FOR el IN ELEMENTS(doc, '.IconNav a')
        RETURN TRIM(el.innerText)
assert:
  text: RETURN T::NOT::EMPTY(@lab.data.query.result)
```

```bash
$ lab mysuite.yaml
```

In order to make testing more modular, you can refer to an existing script in suites:

```yaml
query:
  ref: ../myscript.fql
assert:
  text: RETURN T::NOT::EMPTY(@lab.data.query.result)
```

## Files resolutions

``lab`` supports multiple file locations:

- file:
- git+http:
- git+https:

## Static files serving

``lab`` has an ability to server static files that can be used by your scripts.

```bash
	lab --cdn=./website tests/
```

Which can be access via ``@lab.cdn.DIR_NAME``

```yaml
query:
  text: |
    LET page = DOCUMENT(@lab.cdn.website, { driver: "cdp" })
    
    RETURN page.innerHTML
assert:
  text: RETURN T::NOT::EMPTY(@lab.data.query.result)
```

You can define multiple cdn endpoints pointing to different directories:

```bash
	lab  --cdn=./app_1 --cdn=./app_2 tests/
```

Additionally, you can give them custom names:

```bash
	lab --cdn=./app_1@sales --cdn=./app_2@marketing tests/
```

## Remote Ferret runtime
By default, ``lab`` uses built-in version of Ferret to execute scripts, but it also can use remote versions as well.

- http, https
- bin

### HTTP(S) runtime
HTTP based runtime is used by sending POST requests that contain an object with the following fields:
- query
- params

### External binary runtime
Custom binary runtime is used by using Ferret CLI's interface. 

## Usage

```bash
NAME:
   lab - run FQL test scripts

USAGE:
   lab [global options] [files...]

DESCRIPTION:
   Ferret test runner

COMMANDS:
   version  Show Lab version
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --files value, -f value            location of FQL script files to run [$LAB_FILES]
   --timeout value, -t value          test timeout in seconds (default: 30) [$LAB_TIMEOUT]
   --cdp value                        Chrome DevTools Protocol address (default: "http://127.0.0.1:9222") [$LAB_CDP]
   --reporter value                   reporter (console, simple) (default: "console") [$LAB_REPORTER]
   --runtime value, -r value          url to remote Ferret runtime (http, https or bin) [$LAB_RUNTIME]
   --runtime-param value, --rp value  params for remote Ferret runtime (--runtime-param=headers:{"KeyId": "abcd"} --runtime-param=path:"/ferret" }) [$LAB_RUNTIME_PARAM]
   --concurrency value, -c value      number of multiple tests to run at a time (default: 1) [$LAB_CONCURRENCY]
   --times value                      number of times to run each test (default: 1) [$LAB_TIMES]
   --attempts value, -a value         number of times to re-run failed tests (default: 1) [$LAB_ATTEMPTS]
   --times-interval value             interval between test cycles in seconds (default: 0) [$LAB_TIMES_INTERVAL]
   --cdn value                        file or directory to serve via HTTP (./dir as default or ./dir@name with alias) [$LAB_CDN]
   --param value, -p value            query parameter (--param=foo:"bar", --param=id:1) [$LAB_PARAM]
   --wait value, -w value             tests and waits on the availability of remote resources (--wait http://127.0.0.1:9222/json/version --wait postgres://locahost:5432/mydb) [$LAB_WAIT]
   --wait-timeout value, --wt value   wait timeout in seconds (default: 5) [$LAB_WAIT_TIMEOUT]
   --wait-attempts value              wait attempts (default: 5) [$LAB_WAIT_ATTEMPTS]
   --help, -h                         show help (default: false)
```
