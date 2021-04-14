package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func RequestParamsError(g *gin.Context, message string, err error) {
	g.JSON(http.StatusBadRequest, gin.H{"message": message, "error": err.Error()})
	g.Abort()
}
