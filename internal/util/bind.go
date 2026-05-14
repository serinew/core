package util

import (
	"github.com/gin-gonic/gin"
	"github.com/serinew/core/internal/types"
)

// BindJSON forwards to types (validator / error formatting live in types.init).
func BindJSON(c *gin.Context, ptr any) bool { return types.BindJSON(c, ptr) }

// Guard is an alias for BindJSON.
func Guard(c *gin.Context, ptr any) bool { return types.Guard(c, ptr) }

func BindCreate[T any](c *gin.Context) *T { return types.BindCreate[T](c) }
func BindUpdate[T any](c *gin.Context) *T { return types.BindUpdate[T](c) }
func IsCreate[T any](c *gin.Context) *T   { return types.IsCreate[T](c) }
func IsUpdate[T any](c *gin.Context) *T   { return types.IsUpdate[T](c) }

func FetchListQuery(c *gin.Context) *types.ListQuery { return types.FetchListQuery(c) }
func FetchParams(c *gin.Context) *types.ListQuery    { return types.FetchParams(c) }

func Require(c *gin.Context, ok bool, msg string) bool { return types.Require(c, ok, msg) }
