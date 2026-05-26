package usecase

const (
	defaultPageLimit      = 20
	maxPageLimit          = 100
	maxBlockCount         = 100
	maxMediaFileSize      = 10 * 1024 * 1024
	maxTierNameLen        = 80
	maxTierDescLen        = 500
	minDonationAmount     = 100
	maxDonationAmount     = 1_000_000
	maxDonationMessageLen = 500
	defaultCurrency       = "RUB"
)

func normalizePage(limit int32, offset int32) (int32, int32, error) {
	if limit < 0 || limit > maxPageLimit {
		return 0, 0, ErrInvalidLimit
	}
	if offset < 0 {
		return 0, 0, ErrInvalidOffset
	}
	if limit == 0 {
		limit = defaultPageLimit
	}

	return limit, offset, nil
}
