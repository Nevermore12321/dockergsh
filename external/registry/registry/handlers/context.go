package handlers

import (
	"context"
	"github.com/Nevermore12321/dockergsh/external/registry/registry/api/errcode"
	v1 "github.com/Nevermore12321/dockergsh/external/registry/registry/api/v1"
)

type Context struct {
	*App
	context.Context
	urlBuilder *v1.URLBuilder
	Errors     errcode.Errors
}
