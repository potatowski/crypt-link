package domain

import (
	"fmt"
	"regexp"
)

// MaxEncryptedLength is the maximum accepted length, in bytes, of the
// encrypted payload.
const MaxEncryptedLength = 48 * 1024

var uuidV4Pattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-4[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

// ValidateID checks that id is a UUID v4.
func ValidateID(id string) error {
	if !uuidV4Pattern.MatchString(id) {
		return fmt.Errorf("%w: id must be a UUID v4", ErrInvalidInput)
	}
	return nil
}

// ValidateEncrypted checks that the encrypted payload is present and within size limits.
func ValidateEncrypted(encrypted string) error {
	if encrypted == "" {
		return fmt.Errorf("%w: encrypted content must not be empty", ErrInvalidInput)
	}
	if len(encrypted) > MaxEncryptedLength {
		return fmt.Errorf("%w: encrypted content exceeds %d bytes", ErrInvalidInput, MaxEncryptedLength)
	}
	return nil
}
