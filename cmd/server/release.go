//go:build release
// +build release

package main

import "github.com/gin-gonic/gin"

func init() {
	gin.SetMode(gin.ReleaseMode)
}
