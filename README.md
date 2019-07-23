# AJAX Detector

This tool utilises [chromedp](http://github.com/chromedp/chromedp), an automated browser which supports [Chrome DevTools Protocol](https://chromedevtools.github.io/devtools-protocol/). It aims to enable users to profile a web pages' dependencies using network capability of the DevTools. The dependencies in this context means any fetch or XHR request emitted from the page, which most of the time means other services in which the pages relies on to.

## Installation

To use this tool, after cloning it in your machine, run the `make build` command on the clone directory, it basically will run a `go build` command, and the binary will be built on the `bin/` directory of the repository.

## Usage

There are couple of ways you can run this tool. If you want to run against a single web page, Simply by passing the page URL as a command will suffice:

```text
./page-profile www.tiket.com
```

However, if you'd like to profile several web pages, using the previous snippet multiple times may be cumbersome. In order to ease this, there's a mechanism to load a configuration file by using the `--config-path` file (or just `-c` for short). Simply provide the path to your configuration file as the flag value. **NOTE:** using this flag will ignore the argument in the previous example.

```text
./page-profile -c path/to/config.toml
```

On top of the basic usage, you can use these following flags as well.

```text
Flags:
  -h, --help                 help for page-profile
  -o, --output-path string   Specify directory Path path for output (default "output.txt")
  -t, --timeout int          Set timeout for the execution, in seconds (default 15)
```
