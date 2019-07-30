# AJAX Detector

This tool utilises [chromedp](http://github.com/chromedp/chromedp), an automated browser which supports [Chrome DevTools Protocol](https://chromedevtools.github.io/devtools-protocol/). It aims to enable users to profile a web pages' dependencies using network capability of the DevTools. The dependencies in this context means any fetch or XHR request emitted from the page, which most of the time means other services in which the pages relies on to.

## Installation

To use this tool, after cloning it in your machine, run the `make build` command on the clone directory, it basically will run a `go build` command, and the binary will be built on the `bin/` directory of the repository.

## Usage

There are couple of ways you can run this tool. One is by passing the page URLs as a command argument(s):

```text
./page-profile https://www.tiket.com https://www.tiket.com/pesawat
```

However, there's also a mechanism to load a configuration file by using the `--config` file (or just `-c` for short). Simply provide the path to your configuration file as the flag value. **NOTE:** using this flag will ignore the argument(s) in the previous example. The configuration file is formatted like the following:

```toml
[[pages]]
name = "Tiket.com - Main landing page"
url = "https://www.tiket.com"

[[pages]]
name = "Tiket.com - Pesawat landing page"
url = "https://www.tiket.com/pesawat"

[[pages]]
name = "Tiket.com - Hotel landing page"
url = "https://www.tiket.com/hotel"
```

And to use the configuration file run:

```text
./page-profile -c path/to/config.toml
```

Other than that, you can use these following flags as well.

```text
Flags:
  -c, --config string   Path to configuration file (default "config.toml")
  -h, --help            help for page-profile
  -o, --output string   Specify directory Path path for output (default "output.csv")
  -t, --timeout int     Set timeout for the execution, in seconds (default 15)
```
