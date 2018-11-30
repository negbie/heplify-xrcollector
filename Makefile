NAME?=heplify-xrcollector

all:
	go build -ldflags "-s -w"  -o $(NAME) *.go
