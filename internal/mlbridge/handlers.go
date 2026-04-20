package mlbridge

import (
	"net/http"

	"github.com/emekachisom/kolowise/internal/mlclient"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	ML *mlclient.Client
}

func NewHandler(ml *mlclient.Client) *Handler {
	return &Handler{ML: ml}
}

func (h *Handler) PredictCategory(c *gin.Context) {
	var req mlclient.CategoryPredictionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.ML.PredictCategory(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"prediction": resp})
}

func (h *Handler) PredictSafeToSave(c *gin.Context) {
	var req mlclient.SafeToSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.ML.PredictSafeToSave(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"prediction": resp})
}
