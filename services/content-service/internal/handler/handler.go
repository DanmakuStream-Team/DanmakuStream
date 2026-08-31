package handler

import (
	"errors"
	"net/http"
	"strconv"

	"danmakustream/content-service/internal/config"
	"danmakustream/content-service/internal/logic"
	"danmakustream/content-service/internal/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	DB     *gorm.DB
	Logic  *logic.Service
	Config config.Config
}

func pageParams(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func uintParam(c *gin.Context, name string) (uint, bool) {
	value, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || value == 0 {
		response.Error(c, http.StatusBadRequest, 40000, "invalid "+name)
		return 0, false
	}
	return uint(value), true
}

func writeLogicError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, logic.ErrNotFound):
		response.Error(c, http.StatusNotFound, 40401, "resource not found")
	case errors.Is(err, logic.ErrForbidden):
		response.Error(c, http.StatusForbidden, 40301, "resource ownership required")
	case errors.Is(err, logic.ErrConflict):
		response.Error(c, http.StatusConflict, 40901, "resource state conflict")
	case errors.Is(err, logic.ErrInvalid):
		response.Error(c, http.StatusBadRequest, 40006, "invalid query")
	default:
		response.Error(c, http.StatusInternalServerError, 50000, "internal server error")
	}
}
