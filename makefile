go-nativefier: *.go
	go build -o $(GOPATH)/bin/go-nativefier
	go-nativefier $(url) --title="$(title)" --output="$(output)/$(title)" $(verbose) $(debug)

race: *.go
	go build -o $(GOPATH)/bin/go-nativefier --race
	clear

build: *.go
	go build -o $(GOPATH)/bin/go-nativefier