package handlers

import "context"

type Context struct {
	*App
	context.Context
}
