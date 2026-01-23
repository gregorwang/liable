package services

import (
	"comment-review-platform/internal/config"
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/repository"
	"comment-review-platform/pkg/r2"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	avatarMaxSize           = 1 << 20
	avatarPreviewExpiration = 24 * time.Hour
)

var (
	ErrAvatarTooLarge         = errors.New("avatar too large")
	ErrUnsupportedAvatarType  = errors.New("unsupported avatar type")
	ErrAvatarStorageMissing   = errors.New("R2 storage is not configured")
	ErrAvatarFileInvalid      = errors.New("invalid avatar file")
)

type ProfileService struct {
	userRepo          *repository.UserRepository
	permissionService *PermissionService
	r2                *r2.R2Service
	avatarPrefix      string
}

func NewProfileService() *ProfileService {
	r2Service, err := r2.NewR2Service()
	if err != nil {
		r2Service = nil
	}

	return &ProfileService{
		userRepo:          repository.NewUserRepository(),
		permissionService: NewPermissionService(),
		r2:                r2Service,
		avatarPrefix:      normalizeAvatarPrefix(config.AppConfig.R2AvatarPathPrefix),
	}
}

func (s *ProfileService) GetProfile(userID int) (*models.User, []string, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, nil, err
	}
	s.decorateAvatar(user)

	permissions, err := s.permissionService.GetUserPermissions(userID)
	if err != nil {
		return nil, nil, err
	}
	return user, permissions, nil
}

func (s *ProfileService) UpdateProfile(userID int, req models.UpdateProfileRequest) error {
	gender := normalizeOptionalString(req.Gender)
	signature := normalizeOptionalString(req.Signature)
	return s.userRepo.UpdateProfile(userID, gender, signature)
}

func (s *ProfileService) UpdateSystemProfile(userID int, req models.UpdateSystemProfileRequest) error {
	officeLocation := normalizeOptionalString(req.OfficeLocation)
	department := normalizeOptionalString(req.Department)
	school := normalizeOptionalString(req.School)
	company := normalizeOptionalString(req.Company)
	directManager := normalizeOptionalString(req.DirectManager)
	return s.userRepo.UpdateSystemProfile(userID, officeLocation, department, school, company, directManager)
}

func (s *ProfileService) UpdateAvatar(userID int, fileHeader *multipart.FileHeader) error {
	if fileHeader == nil {
		return ErrAvatarFileInvalid
	}
	if fileHeader.Size > avatarMaxSize {
		return ErrAvatarTooLarge
	}
	if s.r2 == nil {
		return ErrAvatarStorageMissing
	}

	file, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := readAvatarLimited(file, avatarMaxSize)
	if err != nil {
		return err
	}

	contentType := http.DetectContentType(data)
	if !isAllowedAvatarType(contentType) {
		return ErrUnsupportedAvatarType
	}

	ext := strings.ToLower(path.Ext(fileHeader.Filename))
	if ext == "" {
		ext = extensionForAvatarContentType(contentType)
	}
	if ext == "" {
		ext = ".png"
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	key := fmt.Sprintf("%s%d/%s%s", s.avatarPrefix, userID, uuid.NewString(), ext)
	if err := s.r2.UploadObject(key, data, contentType); err != nil {
		return err
	}

	return s.userRepo.UpdateAvatarKey(userID, &key)
}

func (s *ProfileService) decorateAvatar(user *models.User) {
	if user == nil || user.AvatarKey == nil || *user.AvatarKey == "" {
		return
	}
	if s.r2 == nil {
		return
	}
	if url, err := s.r2.GeneratePresignedURL(*user.AvatarKey, avatarPreviewExpiration); err == nil {
		user.AvatarURL = &url
	}
}

func normalizeAvatarPrefix(prefix string) string {
	cleaned := strings.TrimSpace(prefix)
	if cleaned == "" {
		cleaned = "user-avatars/"
	}
	cleaned = strings.TrimPrefix(cleaned, "/")
	if !strings.HasSuffix(cleaned, "/") {
		cleaned += "/"
	}
	return cleaned
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func readAvatarLimited(reader io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, ErrAvatarTooLarge
	}
	return data, nil
}

func isAllowedAvatarType(contentType string) bool {
	switch contentType {
	case "image/png", "image/jpeg", "image/webp":
		return true
	default:
		return false
	}
}

func extensionForAvatarContentType(contentType string) string {
	extensions, err := mime.ExtensionsByType(contentType)
	if err != nil || len(extensions) == 0 {
		return ""
	}
	return extensions[0]
}
