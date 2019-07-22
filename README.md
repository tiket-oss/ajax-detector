# AJAX Detector

This tool utilises [chromedp](http://github.com/chromedp/chromedp), an automated browser which supports [Chrome DevTools Protocol](https://chromedevtools.github.io/devtools-protocol/). It aims to enable users to profile a web pages' dependencies using network capability of the DevTools. The dependencies in this context means any fetch or XHR request emitted from the page, which most of the time means other services in which the pages relies on to.

## Installation

This project is distributed as a [Go Module](https://github.com/golang/go/wiki/Modules). To use this tool, after cloning it in your machine, run the ```make build``` command on the clone directory, and the binary will be built on the `bin/` directory of the repository.

## Usage

Run the compiled binary and provide the target page and output file via `-page-url` and `-file-path` respectively, so the command will look like something along this line: ```./bin/page-profiler -page-url "www.tiket.com" -file-path "~/results/output.txt"
