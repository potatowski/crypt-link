package domain

import (
	"fmt"
	"regexp"
	"time"
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

// DefaultExpiry is used when the client does not choose an expiry.
const DefaultExpiry = 7 * 24 * time.Hour

// allowedExpirySeconds lists the expiry durations, in seconds, a client may choose.
var allowedExpirySeconds = map[int64]bool{
	3600:   true, // 1 hour
	86400:  true, // 1 day
	604800: true, // 7 days
}

// ExpiryFromSeconds converts a client-provided expiry in seconds into a
// duration. Zero means "not provided" and yields DefaultExpiry; any value
// not in the allowed set is rejected with ErrInvalidInput.
func ExpiryFromSeconds(seconds int64) (time.Duration, error) {
	if seconds == 0 {
		return DefaultExpiry, nil
	}
	if !allowedExpirySeconds[seconds] {
		return 0, fmt.Errorf("%w: expiresIn must be one of 3600, 86400 or 604800 seconds", ErrInvalidInput)
	}
	return time.Duration(seconds) * time.Second, nil
}
