package handlers

import (
	"github.com/dienulhaq/go-driving-course-management/dto"
	"github.com/dienulhaq/go-driving-course-management/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

// CreateTrainer godoc
// @Summary Create trainer
// @Description Create trainer. Requires an ACTIVE ADMIN account.
// @Tags Admin Trainers
// @Accept json
// @Produce json
// @Param request body dto.CreateTrainerRequest true "Request body"
// @Success 201 {object} dto.TrainerAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 422 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/trainers [post]
func (h *AdminHandler) CreateTrainer(c *gin.Context) {
	var request dto.CreateTrainerRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request")
		return
	}
	data, err := h.service.CreateTrainer(c.Request.Context(), request)
	if err != nil {
		adminFailure(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "trainer created successfully", data)
}

// ListTrainers godoc
// @Summary List trainers
// @Description List trainers. Requires an ACTIVE ADMIN account.
// @Tags Admin Trainers
// @Produce json
// @Success 200 {object} dto.TrainerListAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/trainers [get]
func (h *AdminHandler) ListTrainers(c *gin.Context) {
	data, err := h.service.ListTrainers(c.Request.Context())
	if err != nil {
		adminFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", data)
}

// GetTrainer godoc
// @Summary Get trainer
// @Description Get trainer. Requires an ACTIVE ADMIN account.
// @Tags Admin Trainers
// @Produce json
// @Param id path int true "Resource ID"
// @Success 200 {object} dto.TrainerAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/trainers/{id} [get]
func (h *AdminHandler) GetTrainer(c *gin.Context) {
	id, ok := adminID(c, "id")
	if !ok {
		return
	}
	data, err := h.service.GetTrainer(c.Request.Context(), id)
	if err != nil {
		adminFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", data)
}

// UpdateTrainer godoc
// @Summary Update trainer
// @Description Update trainer. Requires an ACTIVE ADMIN account.
// @Tags Admin Trainers
// @Accept json
// @Produce json
// @Param id path int true "Resource ID"
// @Param request body dto.UpdateTrainerRequest true "Request body"
// @Success 200 {object} dto.TrainerAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 422 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/trainers/{id} [put]
func (h *AdminHandler) UpdateTrainer(c *gin.Context) {
	id, ok := adminID(c, "id")
	if !ok {
		return
	}
	var request dto.UpdateTrainerRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request")
		return
	}
	data, err := h.service.UpdateTrainer(c.Request.Context(), id, request)
	if err != nil {
		adminFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "trainer updated successfully", data)
}

// UpdateTrainerStatus godoc
// @Summary Update trainer status
// @Description Update trainer status. Requires an ACTIVE ADMIN account.
// @Tags Admin Trainers
// @Accept json
// @Produce json
// @Param id path int true "Resource ID"
// @Param request body dto.UpdateStatusRequest true "Request body"
// @Success 200 {object} dto.TrainerAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/trainers/{id}/status [patch]
func (h *AdminHandler) UpdateTrainerStatus(c *gin.Context) {
	id, ok := adminID(c, "id")
	if !ok {
		return
	}
	var request dto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request")
		return
	}
	data, err := h.service.UpdateTrainerStatus(c.Request.Context(), id, request.Status)
	if err != nil {
		adminFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "trainer status updated successfully", data)
}

// ListStudents godoc
// @Summary List students
// @Description List students. Requires an ACTIVE ADMIN account.
// @Tags Admin Students
// @Produce json
// @Success 200 {object} dto.StudentListAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/students [get]
func (h *AdminHandler) ListStudents(c *gin.Context) {
	data, err := h.service.ListStudents(c.Request.Context())
	if err != nil {
		adminFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", data)
}

// GetStudent godoc
// @Summary Get student
// @Description Get student. Requires an ACTIVE ADMIN account.
// @Tags Admin Students
// @Produce json
// @Param id path int true "Resource ID"
// @Success 200 {object} dto.StudentAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/students/{id} [get]
func (h *AdminHandler) GetStudent(c *gin.Context) {
	id, ok := adminID(c, "id")
	if !ok {
		return
	}
	data, err := h.service.GetStudent(c.Request.Context(), id)
	if err != nil {
		adminFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", data)
}

// UpdateStudentStatus godoc
// @Summary Update student status
// @Description Update student status. Requires an ACTIVE ADMIN account.
// @Tags Admin Students
// @Accept json
// @Produce json
// @Param id path int true "Resource ID"
// @Param request body dto.UpdateStatusRequest true "Request body"
// @Success 200 {object} dto.StudentAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/students/{id}/status [patch]
func (h *AdminHandler) UpdateStudentStatus(c *gin.Context) {
	id, ok := adminID(c, "id")
	if !ok {
		return
	}
	var request dto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request")
		return
	}
	data, err := h.service.UpdateStudentStatus(c.Request.Context(), id, request.Status)
	if err != nil {
		adminFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "student status updated successfully", data)
}

// CreatePackage godoc
// @Summary Create course package
// @Description Create course package. Requires an ACTIVE ADMIN account.
// @Tags Admin Packages
// @Accept json
// @Produce json
// @Param request body dto.UpsertPackageRequest true "Request body"
// @Success 201 {object} dto.PackageAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 422 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/packages [post]
func (h *AdminHandler) CreatePackage(c *gin.Context) {
	var request dto.UpsertPackageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request")
		return
	}
	data, err := h.service.CreatePackage(c.Request.Context(), request)
	if err != nil {
		adminFailure(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "course package created successfully", data)
}

// ListPackages godoc
// @Summary List course packages
// @Description List course packages. Requires an ACTIVE ADMIN account.
// @Tags Admin Packages
// @Produce json
// @Success 200 {object} dto.PackageListAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/packages [get]
func (h *AdminHandler) ListPackages(c *gin.Context) {
	data, err := h.service.ListPackages(c.Request.Context())
	if err != nil {
		adminFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", data)
}

// GetPackage godoc
// @Summary Get course package
// @Description Get course package. Requires an ACTIVE ADMIN account.
// @Tags Admin Packages
// @Produce json
// @Param id path int true "Resource ID"
// @Success 200 {object} dto.PackageAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/packages/{id} [get]
func (h *AdminHandler) GetPackage(c *gin.Context) {
	id, ok := adminID(c, "id")
	if !ok {
		return
	}
	data, err := h.service.GetPackage(c.Request.Context(), id)
	if err != nil {
		adminFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", data)
}

// UpdatePackage godoc
// @Summary Update course package
// @Description Update course package. Requires an ACTIVE ADMIN account.
// @Tags Admin Packages
// @Accept json
// @Produce json
// @Param id path int true "Resource ID"
// @Param request body dto.UpsertPackageRequest true "Request body"
// @Success 200 {object} dto.PackageAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 422 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/packages/{id} [put]
func (h *AdminHandler) UpdatePackage(c *gin.Context) {
	id, ok := adminID(c, "id")
	if !ok {
		return
	}
	var request dto.UpsertPackageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request")
		return
	}
	data, err := h.service.UpdatePackage(c.Request.Context(), id, request)
	if err != nil {
		adminFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "course package updated successfully", data)
}

// UpdatePackageStatus godoc
// @Summary Update package status
// @Description Update package status. Requires an ACTIVE ADMIN account.
// @Tags Admin Packages
// @Accept json
// @Produce json
// @Param id path int true "Resource ID"
// @Param request body dto.UpdateStatusRequest true "Request body"
// @Success 200 {object} dto.PackageAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/packages/{id}/status [patch]
func (h *AdminHandler) UpdatePackageStatus(c *gin.Context) {
	id, ok := adminID(c, "id")
	if !ok {
		return
	}
	var request dto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request")
		return
	}
	data, err := h.service.UpdatePackageStatus(c.Request.Context(), id, request.Status)
	if err != nil {
		adminFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "package status updated successfully", data)
}

// CreateMaterial godoc
// @Summary Create curriculum material
// @Description Create curriculum material. Requires an ACTIVE ADMIN account.
// @Tags Admin Materials
// @Accept json
// @Produce json
// @Param request body dto.UpsertMaterialRequest true "Request body"
// @Success 201 {object} dto.MaterialAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 422 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/materials [post]
func (h *AdminHandler) CreateMaterial(c *gin.Context) {
	var request dto.UpsertMaterialRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request")
		return
	}
	data, err := h.service.CreateMaterial(c.Request.Context(), request)
	if err != nil {
		adminFailure(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "curriculum material created successfully", data)
}

// ListMaterials godoc
// @Summary List curriculum materials
// @Description List curriculum materials. Requires an ACTIVE ADMIN account.
// @Tags Admin Materials
// @Produce json
// @Success 200 {object} dto.MaterialListAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/materials [get]
func (h *AdminHandler) ListMaterials(c *gin.Context) {
	data, err := h.service.ListMaterials(c.Request.Context())
	if err != nil {
		adminFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", data)
}

// GetMaterial godoc
// @Summary Get curriculum material
// @Description Get curriculum material. Requires an ACTIVE ADMIN account.
// @Tags Admin Materials
// @Produce json
// @Param id path int true "Resource ID"
// @Success 200 {object} dto.MaterialAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/materials/{id} [get]
func (h *AdminHandler) GetMaterial(c *gin.Context) {
	id, ok := adminID(c, "id")
	if !ok {
		return
	}
	data, err := h.service.GetMaterial(c.Request.Context(), id)
	if err != nil {
		adminFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", data)
}

// UpdateMaterial godoc
// @Summary Update curriculum material
// @Description Update curriculum material. Requires an ACTIVE ADMIN account.
// @Tags Admin Materials
// @Accept json
// @Produce json
// @Param id path int true "Resource ID"
// @Param request body dto.UpsertMaterialRequest true "Request body"
// @Success 200 {object} dto.MaterialAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 422 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/materials/{id} [put]
func (h *AdminHandler) UpdateMaterial(c *gin.Context) {
	id, ok := adminID(c, "id")
	if !ok {
		return
	}
	var request dto.UpsertMaterialRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request")
		return
	}
	data, err := h.service.UpdateMaterial(c.Request.Context(), id, request)
	if err != nil {
		adminFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "curriculum material updated successfully", data)
}

// UpdateMaterialStatus godoc
// @Summary Update material status
// @Description Update material status. Requires an ACTIVE ADMIN account.
// @Tags Admin Materials
// @Accept json
// @Produce json
// @Param id path int true "Resource ID"
// @Param request body dto.UpdateStatusRequest true "Request body"
// @Success 200 {object} dto.MaterialAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/materials/{id}/status [patch]
func (h *AdminHandler) UpdateMaterialStatus(c *gin.Context) {
	id, ok := adminID(c, "id")
	if !ok {
		return
	}
	var request dto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request")
		return
	}
	data, err := h.service.UpdateMaterialStatus(c.Request.Context(), id, request.Status)
	if err != nil {
		adminFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "material status updated successfully", data)
}

// CreateSubMaterial godoc
// @Summary Create sub-material
// @Description Create sub-material. Requires an ACTIVE ADMIN account.
// @Tags Admin Sub Materials
// @Accept json
// @Produce json
// @Param material_id path int true "Resource ID"
// @Param request body dto.UpsertSubMaterialRequest true "Request body"
// @Success 201 {object} dto.SubMaterialAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 422 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/materials/{material_id}/sub-materials [post]
func (h *AdminHandler) CreateSubMaterial(c *gin.Context) {
	id, ok := adminID(c, "id")
	if !ok {
		return
	}
	var request dto.UpsertSubMaterialRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request")
		return
	}
	data, err := h.service.CreateSubMaterial(c.Request.Context(), id, request)
	if err != nil {
		adminFailure(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "sub-material created successfully", data)
}

// ListSubMaterials godoc
// @Summary List sub-materials
// @Description List sub-materials. Requires an ACTIVE ADMIN account.
// @Tags Admin Sub Materials
// @Produce json
// @Param material_id path int true "Resource ID"
// @Success 200 {object} dto.SubMaterialListAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/materials/{material_id}/sub-materials [get]
func (h *AdminHandler) ListSubMaterials(c *gin.Context) {
	id, ok := adminID(c, "id")
	if !ok {
		return
	}
	data, err := h.service.ListSubMaterials(c.Request.Context(), id)
	if err != nil {
		adminFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", data)
}

// GetSubMaterial godoc
// @Summary Get sub-material
// @Description Get sub-material. Requires an ACTIVE ADMIN account.
// @Tags Admin Sub Materials
// @Produce json
// @Param id path int true "Resource ID"
// @Success 200 {object} dto.SubMaterialAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/sub-materials/{id} [get]
func (h *AdminHandler) GetSubMaterial(c *gin.Context) {
	id, ok := adminID(c, "id")
	if !ok {
		return
	}
	data, err := h.service.GetSubMaterial(c.Request.Context(), id)
	if err != nil {
		adminFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", data)
}

// UpdateSubMaterial godoc
// @Summary Update sub-material
// @Description Update sub-material. Requires an ACTIVE ADMIN account.
// @Tags Admin Sub Materials
// @Accept json
// @Produce json
// @Param id path int true "Resource ID"
// @Param request body dto.UpsertSubMaterialRequest true "Request body"
// @Success 200 {object} dto.SubMaterialAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 422 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/sub-materials/{id} [put]
func (h *AdminHandler) UpdateSubMaterial(c *gin.Context) {
	id, ok := adminID(c, "id")
	if !ok {
		return
	}
	var request dto.UpsertSubMaterialRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request")
		return
	}
	data, err := h.service.UpdateSubMaterial(c.Request.Context(), id, request)
	if err != nil {
		adminFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "sub-material updated successfully", data)
}

// UpdateSubMaterialStatus godoc
// @Summary Update sub-material status
// @Description Update sub-material status. Requires an ACTIVE ADMIN account.
// @Tags Admin Sub Materials
// @Accept json
// @Produce json
// @Param id path int true "Resource ID"
// @Param request body dto.UpdateStatusRequest true "Request body"
// @Success 200 {object} dto.SubMaterialAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/sub-materials/{id}/status [patch]
func (h *AdminHandler) UpdateSubMaterialStatus(c *gin.Context) {
	id, ok := adminID(c, "id")
	if !ok {
		return
	}
	var request dto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request")
		return
	}
	data, err := h.service.UpdateSubMaterialStatus(c.Request.Context(), id, request.Status)
	if err != nil {
		adminFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "sub-material status updated successfully", data)
}
