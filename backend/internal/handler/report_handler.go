package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/internal/model"
)

var actionLabels = map[string]string{
	"delete_listing": "Xóa tin đăng",
	"block_user":     "Khóa người dùng",
	"warn_user":      "Cảnh cáo",
	"dismiss":        "Bỏ qua báo cáo",
}

var targetLabels = map[string]string{
	"listing": "tin đăng",
	"user":    "người dùng",
	"rating":  "đánh giá",
}

type ReportHandler struct {
	reportService  ReportServiceInterface
	notifService   NotificationServiceInterface
	listingService ListingServiceInterface
	adminService   AdminServiceInterface
}

func NewReportHandler(reportService ReportServiceInterface, notifService NotificationServiceInterface, listingService ListingServiceInterface, adminService AdminServiceInterface) *ReportHandler {
	return &ReportHandler{
		reportService:  reportService,
		notifService:   notifService,
		listingService: listingService,
		adminService:   adminService,
	}
}

func (h *ReportHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")

	var req model.CreateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_type (listing/user/rating), target_id, and reason are required"})
		return
	}

	report, err := h.reportService.Create(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create report"})
		return
	}

	c.JSON(http.StatusCreated, report)
}

func (h *ReportHandler) ListPending(c *gin.Context) {
	page, limit := parsePagination(c, 20)
	status := c.DefaultQuery("status", "pending")

	// Validate status
	if status != "pending" && status != "resolved" && status != "dismissed" && status != "all" {
		status = "pending"
	}

	reports, total, err := h.reportService.ListByStatus(c.Request.Context(), status, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list reports"})
		return
	}

	totalPages := (total + limit - 1) / limit
	c.JSON(http.StatusOK, model.PaginatedResponse{
		Data: reports, Total: total, Page: page, Limit: limit, TotalPages: totalPages,
	})
}

func (h *ReportHandler) Resolve(c *gin.Context) {
	adminID := c.GetString("user_id")
	reportID := c.Param("id")

	var req model.ResolveReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "admin_action is required"})
		return
	}

	report, err := h.reportService.Resolve(c.Request.Context(), reportID, adminID, req.AdminAction, req.AdminNote)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve report"})
		return
	}

	// Resolve target owner BEFORE executing action (which may delete the listing)
	targetOwnerID := h.getTargetOwnerID(c.Request.Context(), report)

	// Execute the actual admin action
	if execErr := h.executeAction(c.Request.Context(), report); execErr != nil {
		log.Printf("Failed to execute action %s for report %s: %v", req.AdminAction, reportID, execErr)
	}

	// Send notifications to reporter and target owner (best effort, don't fail the request)
	go h.sendResolveNotifications(report, targetOwnerID)

	c.JSON(http.StatusOK, report)
}

func (h *ReportHandler) executeAction(ctx context.Context, report *model.Report) error {
	if h.adminService == nil {
		return nil
	}
	action := deref(report.AdminAction)
	switch action {
	case "delete_listing":
		if report.TargetType == "listing" {
			return h.adminService.DeleteListing(ctx, report.TargetID)
		}
	case "block_user":
		ownerID := h.getTargetOwnerID(ctx, report)
		if ownerID != "" {
			reason := fmt.Sprintf("Vi phạm: %s", report.Reason)
			if report.AdminNote != nil && *report.AdminNote != "" {
				reason = fmt.Sprintf("%s — %s", reason, *report.AdminNote)
			}
			_, err := h.adminService.BlockUser(ctx, ownerID, reason, "admin")
			return err
		}
	case "warn_user":
		// Send warning notification to the target owner
		if h.notifService != nil {
			ownerID := h.getTargetOwnerID(ctx, report)
			if ownerID != "" {
				noteText := ""
				if report.AdminNote != nil && *report.AdminNote != "" {
					noteText = fmt.Sprintf("\nChi tiết: %s", *report.AdminNote)
				}
				title := "Cảnh cáo vi phạm"
				body := fmt.Sprintf("Bạn đã bị cảnh cáo vì vi phạm: %s. Vui lòng tuân thủ quy định để tránh bị khóa tài khoản.%s", report.Reason, noteText)
				data, _ := json.Marshal(map[string]string{
					"report_id": report.ID,
					"action":    "warn_user",
				})
				_, err := h.notifService.Create(ctx, ownerID, "warning", title, body, data)
				return err
			}
		}
	}
	return nil
}

func (h *ReportHandler) sendResolveNotifications(report *model.Report, targetOwnerID string) {
	if h.notifService == nil {
		return
	}
	ctx := context.Background()
	actionLabel := actionLabels[deref(report.AdminAction)]
	if actionLabel == "" {
		actionLabel = deref(report.AdminAction)
	}
	targetLabel := targetLabels[report.TargetType]
	if targetLabel == "" {
		targetLabel = report.TargetType
	}

	noteText := ""
	if report.AdminNote != nil && *report.AdminNote != "" {
		noteText = fmt.Sprintf("\nGhi chú từ quản trị: %s", *report.AdminNote)
	}

	data, _ := json.Marshal(map[string]string{
		"report_id":    report.ID,
		"target_type":  report.TargetType,
		"target_id":    report.TargetID,
		"admin_action": deref(report.AdminAction),
	})

	// 1. Notify reporter
	reporterTitle := "Báo cáo của bạn đã được xử lý"
	reporterBody := fmt.Sprintf("Báo cáo %s về %s đã được xử lý. Quyết định: %s.%s",
		report.Reason, targetLabel, actionLabel, noteText)

	if _, err := h.notifService.Create(ctx, report.ReporterID, "report_resolved", reporterTitle, reporterBody, data); err != nil {
		log.Printf("Failed to notify reporter %s: %v", report.ReporterID, err)
	}

	// 2. Notify target owner
	if targetOwnerID != "" && targetOwnerID != report.ReporterID {
		ownerTitle := "Nội dung của bạn đã bị xử lý"
		ownerBody := fmt.Sprintf("%s của bạn đã bị báo cáo vi phạm (%s). Quyết định: %s.%s",
			capitalize(targetLabel), report.Reason, actionLabel, noteText)

		if _, err := h.notifService.Create(ctx, targetOwnerID, "report_resolved", ownerTitle, ownerBody, data); err != nil {
			log.Printf("Failed to notify target owner %s: %v", targetOwnerID, err)
		}
	}
}

func (h *ReportHandler) getTargetOwnerID(ctx context.Context, report *model.Report) string {
	switch report.TargetType {
	case "user":
		return report.TargetID
	case "listing":
		listing, err := h.listingService.GetByID(ctx, report.TargetID)
		if err != nil {
			log.Printf("Failed to get listing %s for notification: %v", report.TargetID, err)
			return ""
		}
		return listing.UserID
	default:
		return ""
	}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	// Vietnamese strings: capitalize first rune
	runes := []rune(s)
	if runes[0] >= 'a' && runes[0] <= 'z' {
		runes[0] -= 32
	}
	return string(runes)
}
