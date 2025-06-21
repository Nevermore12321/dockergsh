package actions

import "errors"

var (
	ErrCmdFormat = errors.New("the format of the command you entered is incorrect. Please use -h to check usage")
)
