package handlers

import (
	"designreview/internal/middleware"
	"designreview/internal/models"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// errMsgNoDesigner dipakai semua handler di file ini saat context tidak berisi
// identitas designer (seharusnya tidak terjadi bila RequireDesigner terpasang).
const errMsgNoDesigner = "missing designer identity"

// GetAllProjects: GET /api/v1/projects -> 200 { projects: [...] }
// Hanya project milik designer yang login (docs/api.md:12 — tiap project punya owner).
func (h *Handler) GetAllProjects(c *gin.Context) {
	designerID := c.GetString(middleware.DesignerIDKey)
	if designerID == "" {
		Fail(c, http.StatusUnauthorized, "unauthorized", errMsgNoDesigner, nil)
		return
	}

	// Slice, bukan satu struct: Find(&project) hanya mengisi baris pertama
	// dan membuang sisanya.
	var projects []models.Project
	if err := h.DB.WithContext(c.Request.Context()).
		Where("owner_id = ?", designerID).
		Order("created_at DESC").
		Find(&projects).Error; err != nil {
		Fail(c, http.StatusInternalServerError, "internal", "failed to fetch projects", nil)
		return
	}

	OK(c, "successfully get projects", gin.H{"projects": projects})
}

// ProjectRequest memetakan body POST/PATCH /projects.
// description opsional (docs/api.md:62) — DB punya DEFAULT ''.
// CATATAN: tag binding TIDAK boleh ada spasi setelah koma
// ("required, min=8" salah -> validator " min=8" tidak dikenal).
type ProjectRequest struct {
	Name        string `json:"name" binding:"required,min=2"`
	Description string `json:"description"`
}


func (h *Handler) CreateProject(c *gin.Context) {
	designerID := c.GetString(middleware.DesignerIDKey)
	if designerID == "" {
		Fail(c, http.StatusUnauthorized, "unauthorized", errMsgNoDesigner, nil)
		return
	}

	var req ProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid_input", err.Error(), nil)
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		BadRequest(c, "invalid_input", "name must not be empty", nil)
		return
	}

	
	project := models.Project{
		OwnerID:     designerID,
		Name:        name,
		Description: strings.TrimSpace(req.Description),
	}

	if err := h.DB.WithContext(c.Request.Context()).Create(&project).Error; err != nil {
		Fail(c, http.StatusInternalServerError, "internal", "failed to create project", nil)
		return 
	}

	Created(c, "successfully create project", gin.H{"project": project})
}


// GetProject: GET /api/v1/projects/:projectId
// -> 200 { project, designs: [...] } — tiap design membawa latest_version (docs/api.md:63).
// Project milik designer lain dibalas 404 (bukan 403) supaya keberadaannya tidak bocor.
func (h *Handler) GetProject(c *gin.Context) {
	designerID := c.GetString(middleware.DesignerIDKey)
	if designerID == "" {
		Fail(c, http.StatusUnauthorized, "unauthorized", errMsgNoDesigner, nil)
		return
	}

	project, ok := h.findOwnedProject(c, designerID)
	if !ok {
		return // helper sudah menulis response error
	}

	designs, err := h.listDesigns(c, project.ID)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal", "failed to fetch designs", nil)
		return
	}

	if err := h.attachLatestVersions(c, designs); err != nil {
		Fail(c, http.StatusInternalServerError, "internal", "failed to fetch versions", nil)
		return
	}

	OK(c, "successfully get project", gin.H{"project": project, "designs": designs})
}

// findOwnedProject mengambil project milik designerID, atau menulis response
// 404/500 dan mengembalikan ok=false. Semua jalur gagal ditangani di sini
// supaya GetProject tetap linear.
func (h *Handler) findOwnedProject(c *gin.Context, designerID string) (*models.Project, bool) {
	projectID := c.Param("projectId")

	// Validasi UUID lebih dulu: kalau tidak, Postgres menolak string non-UUID
	// (error 22P02) dan kita salah balas 500, padahal ini cuma id ngawur.
	if _, err := uuid.Parse(projectID); err != nil {
		Fail(c, http.StatusNotFound, "not_found", "project not found", nil)
		return nil, false
	}

	var project models.Project
	err := h.DB.WithContext(c.Request.Context()).
		Where("id = ? AND owner_id = ?", projectID, designerID).
		First(&project).Error
	if err == nil {
		return &project, true
	}

	// First() (beda dari Find()) mengembalikan ErrRecordNotFound -> 404.
	if errors.Is(err, gorm.ErrRecordNotFound) {
		Fail(c, http.StatusNotFound, "not_found", "project not found", nil)
		return nil, false
	}
	Fail(c, http.StatusInternalServerError, "internal", "failed to fetch project", nil)
	return nil, false
}

func (h *Handler) listDesigns(c *gin.Context, projectID string) ([]models.Design, error) {
	var designs []models.Design
	err := h.DB.WithContext(c.Request.Context()).
		Where("project_id = ?", projectID).
		Order("created_at DESC").
		Find(&designs).Error
	return designs, err
}

// attachLatestVersions mengisi LatestVersion tiap design dengan nomor versi
// tertinggi. Mengambil semua versi dalam SATU query (menghindari N+1), lalu
// memilih max per design di memori.
func (h *Handler) attachLatestVersions(c *gin.Context, designs []models.Design) error {
	if len(designs) == 0 {
		return nil
	}

	designIDs := make([]string, len(designs))
	for i, d := range designs {
		designIDs[i] = d.ID
	}

	var versions []models.Version
	if err := h.DB.WithContext(c.Request.Context()).
		Where("design_id IN ?", designIDs).
		Find(&versions).Error; err != nil {
		return err
	}

	latest := make(map[string]models.Version, len(versions))
	for _, v := range versions {
		if cur, ok := latest[v.DesignID]; !ok || v.Number > cur.Number {
			latest[v.DesignID] = v
		}
	}
	for i := range designs {
		if v, ok := latest[designs[i].ID]; ok {
			designs[i].LatestVersion = &models.VersionBrief{ID: v.ID, Number: v.Number}
		}
	}
	return nil
}
