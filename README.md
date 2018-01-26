# go-nativefier

Make any web page a desktop application. Inspired by [nativefier](https://github.com/jiahaog/nativefier). Written in Go. Tiny binaries (5mb). Native web engines.

> Note: only osx is supported in this initial draft.

[nativefier](https://github.com/jiahaog/nativefier) has a few issues I don't like:

* Huge binaries (250mb in some cases)
* Depends on electron toolchain
* Depends on Nodejs toolchain

go-nativefier has no dependencies on Nodejs or electron. It uses native web engines on each respective OS. Single binary distribution. Size is under 10mb.

## Roadmap

* [ ] Native "bundles"  
        * [x] MacOS `.app`
        * [ ] Windows `.exe`
        * [ ] Linux elf binary
* [ ] Icon support  
        * [x] Icon inference
        * [x] Icon conversion
* [ ] App name inference  
* [ ] Avoid "cold starts"  
        * [ ] Cache website contents
        * [ ] Show loading animation instead of blank white screen

## Caveats

* No dev-tools
* Potential incompatibilities between native web engines
* No desktop integration that comes with electron

## Todo

* Bundler interface to abstract creating single file bundles for the different OS's

## Bugs

* Webview appears to not eval injected javascript for webkit on osx