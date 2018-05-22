package main

var (
	config *configure
)

func main() {
	Execute()
}

func init() {
	config = new(configure)
	initLogger()
}
