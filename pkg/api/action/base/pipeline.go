package base

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s baseServer) ListPipeLine(g *gin.Context) {
	g.JSON(http.StatusOK, "")
}
