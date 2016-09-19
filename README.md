##justdoit

![Project status](https://img.shields.io/badge/version-1.0.0-green.svg)
[![Build Status](https://semaphoreci.com/api/v1/joeybloggs/justdoit/branches/master/badge.svg)](https://semaphoreci.com/joeybloggs/justdoit)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-playground/justdoit)](https://goreportcard.com/report/github.com/go-playground/justdoit)
![License](https://img.shields.io/dub/l/vibe-d.svg)

Why another auto-compile daemon
-----------------------
I couldn't find all of the very basic options I needed that worked correctly in
a single package, most are over-stuffed with features... but streaming of logs from your 
app isn't supported for example. I needed something that just worked hence the name 'justdoit'

Installation
----
```go
go get github.com/go-playground/justdoit
```

Usage
-----
```
justdoit -h

Usage of justdoit:
  -build string
    	Command to Build/Compile program (default "go install -v")
  -exclude string
    	Regex of paths to exclude (default "(.git|vendor)$")
  -include string
    	Regex of files to include (default "(.+\\.go|.+\\.c)$")
  -run string
    	Command to run your application
  -watch string
    	Directory to watch for changes (recursive) (default "./")
```

Example
-------
```
justdoit -watch="./" -include="(.+\\.go|.+\\.c)$" -build="go install -v" -run="$GOPATH/bin/myexecutable"
```

Licenses
--------
- [MIT License](https://raw.githubusercontent.com/go-playground/justdoit/master/LICENSE) (MIT), Copyright (c) 2015 Dean Karn
