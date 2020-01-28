package gsmcache

import "log"

const debug = false

// SMCache TODO
type SMCache struct {
	ProjectId    string
	SecretPrefix string
}

func dlog(format string, v ...interface{}) {
	if debug {
		log.Printf(format, v...)
	}
}
