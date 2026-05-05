package domain

type PostMedia struct {
	FileURL     string
	Kind        BlockKind
	ContentType string
	SizeBytes   int64
}
