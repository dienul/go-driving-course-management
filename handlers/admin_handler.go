package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/dienulhaq/go-driving-course-management/services"
	"github.com/dienulhaq/go-driving-course-management/utils"
	"github.com/gin-gonic/gin"
)

type AdminHandler struct{ service *services.AdminService }

func NewAdminHandler(service *services.AdminService) *AdminHandler {
	return &AdminHandler{service: service}
}

func adminID(c *gin.Context, parameter string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(parameter), 10, 64)
	if err != nil || id <= 0 {
		utils.Error(c, http.StatusBadRequest, "invalid resource id")
		return 0, false
	}
	return id, true
}

func adminFailure(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrResourceNotFound):
		utils.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, services.ErrResourceConflict), errors.Is(err, services.ErrEmailExists):
		utils.Error(c, http.StatusConflict, err.Error())
	case errors.Is(err, services.ErrInvalidInput):
		utils.Error(c, http.StatusUnprocessableEntity, err.Error())
	default:
		utils.InternalServerError(c)
	}
}
