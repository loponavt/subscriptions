package handler

import (
	"errors"
	"log"
	"net/http"
	"subscriptions/internal/model"
	"subscriptions/internal/usecase"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "subscriptions/docs"
)

type Handler struct {
	Usecase *usecase.Usecase
}

func New(s *usecase.Usecase) *Handler {
	return &Handler{Usecase: s}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	sub := r.Group("/subscriptions")
	{
		sub.POST("", h.CreateSubscription)
		sub.GET("", h.List)
		sub.GET("/:id", h.Get)
		sub.PUT("/:id", h.Update)
		sub.DELETE("/:id", h.Delete)
	}

	r.GET("/subscriptions/total", h.Total)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// CreateSubscription godoc
// @Summary Create a new subscription
// @Description Create a subscription with service name, price, user ID, start and optional end dates
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body model.SubscriptionReq true "Subscription request body"
// @Success 201 {object} model.Subscription
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions [post]
func (h *Handler) CreateSubscription(c *gin.Context) {
	var req model.SubscriptionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Price < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "price must be a positive integer"})
		return
	}

	startTime, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, expected MM-YYYY"})
		return
	}

	var endTime *time.Time
	if req.EndDate != nil {
		t, err := time.Parse("01-2006", *req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, expected MM-YYYY"})
			return
		}
		if startTime.After(t) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "start_date must be before or equal to end_date"})
			return
		}
		endTime = &t
	}

	sub := &model.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   startTime,
		EndDate:     endTime,
	}

	if err := h.Usecase.CreateSubscription(sub); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("Created subscription with ID: %s", sub.ID.String())

	c.JSON(http.StatusCreated, sub)
}

// Get godoc
// @Summary Get subscription by ID
// @Description Get details of a subscription by its UUID
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID (UUID)"
// @Success 200 {object} model.Subscription
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /subscriptions/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid UUID"})
		return
	}
	sub, err := h.Usecase.GetSubscription(id)
	if err != nil {
		if errors.Is(err, usecase.ErrSubscriptionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, sub)
}

// Update godoc
// @Summary Update subscription by ID
// @Description Update subscription fields by subscription UUID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID (UUID)"
// @Param subscription body model.SubscriptionReq true "Updated subscription data"
// @Success 200 {object} model.Subscription
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid UUID"})
		return
	}
	var subReq model.SubscriptionReq
	if err := c.ShouldBindJSON(&subReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startDate, err := time.Parse("01-2006", subReq.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, expected MM-YYYY"})
		return
	}

	var endDatePtr *time.Time
	if subReq.EndDate != nil {
		endDate, err := time.Parse("01-2006", *subReq.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, expected MM-YYYY"})
			return
		}
		endDatePtr = &endDate
	}

	sub := model.Subscription{
		ID:          id,
		ServiceName: subReq.ServiceName,
		Price:       subReq.Price,
		UserID:      subReq.UserID,
		StartDate:   startDate,
		EndDate:     endDatePtr,
	}

	if err := h.Usecase.UpdateSubscription(&sub); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sub)
}

// Delete godoc
// @Summary Delete subscription by ID
// @Description Delete subscription by UUID
// @Tags subscriptions
// @Param id path string true "Subscription ID (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid UUID"})
		return
	}
	if err := h.Usecase.DeleteSubscription(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// List godoc
// @Summary List subscriptions
// @Description Get all subscriptions optionally filtered by user_id and service_name
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "Filter by user UUID"
// @Param service_name query string false "Filter by service name"
// @Success 200 {array} model.Subscription
// @Failure 500 {object} map[string]string
// @Router /subscriptions [get]
func (h *Handler) List(c *gin.Context) {
	userIDStr := c.Query("user_id")
	serviceName := c.Query("service_name")

	var userID *uuid.UUID
	if userIDStr != "" {
		id, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
			return
		}
		userID = &id
	}

	var svcName *string
	if serviceName != "" {
		svcName = &serviceName
	}

	subs, err := h.Usecase.ListSubscriptions(userID, svcName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, subs)
}

// Total godoc
// @Summary Calculate total subscription cost
// @Description Calculate total cost for a user and optional service within a date range (from, to in MM-YYYY format)
// @Tags subscriptions
// @Produce json
// @Param user_id query string true "User UUID"
// @Param service_name query string false "Service name"
// @Param from query string false "Start period (MM-YYYY)"
// @Param to query string false "End period (MM-YYYY)"
// @Success 200 {object} map[string]int
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/total [get]
func (h *Handler) Total(c *gin.Context) {
	userIDStr := c.Query("user_id")
	serviceName := c.Query("service_name")
	fromStr := c.Query("from")
	toStr := c.Query("to")

	var userID *uuid.UUID
	if userIDStr != "" {
		id, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
			return
		}
		userID = &id
	}

	var svcName *string
	if serviceName != "" {
		svcName = &serviceName
	}

	// Парсим from и to — ожидаем формат "01-2006" (MM-YYYY)
	const dateLayout = "01-2006"
	if fromStr == "" || toStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from and to parameters are required"})
		return
	}

	from, err := time.Parse(dateLayout, fromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from date format, expected MM-YYYY"})
		return
	}

	to, err := time.Parse(dateLayout, toStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to date format, expected MM-YYYY"})
		return
	}

	if from.After(to) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from must be before or equal to to"})
		return
	}

	sum, err := h.Usecase.CalculateTotal(userID, svcName, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"total": sum})
}
