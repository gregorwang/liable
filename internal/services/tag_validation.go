package services

import (
	"comment-review-platform/internal/repository"
	"fmt"
)

func validateTags(tagRepo *repository.TagRepository, scope string, tags []string) error {
	if len(tags) == 0 {
		return nil
	}
	allowed, err := tagRepo.FindActiveNamesByScope(scope)
	if err != nil {
		return err
	}
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, tag := range allowed {
		allowedSet[tag] = struct{}{}
	}
	for _, tag := range tags {
		if _, ok := allowedSet[tag]; !ok {
			return fmt.Errorf("invalid tag: %s", tag)
		}
	}
	return nil
}
