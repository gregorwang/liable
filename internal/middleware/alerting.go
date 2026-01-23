package middleware

import "comment-review-platform/internal/services"

var alertService *services.AlertService

// InitAlertService initializes the global alert service.
func InitAlertService(service *services.AlertService) {
	alertService = service
}

func getAlertService() *services.AlertService {
	return alertService
}
