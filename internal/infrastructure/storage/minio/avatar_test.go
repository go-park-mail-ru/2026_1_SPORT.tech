package minioadapter

import "testing"

func TestAvatarExtension(t *testing.T) {
	tests := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
		"text/plain": "",
	}

	for contentType, expected := range tests {
		if got := avatarExtension(contentType, "file"); got != expected {
			t.Fatalf("unexpected extension for %s: got %q, expect %q", contentType, got, expected)
		}
	}
}

func TestRandomObjectIDNotEmpty(t *testing.T) {
	if value := randomObjectID(); value == "" {
		t.Fatal("expected non-empty object id")
	}
}

func TestRandomObjectIDLengthAndUniqueness(t *testing.T) {
	first := randomObjectID()
	second := randomObjectID()

	if len(first) != 32 || len(second) != 32 {
		t.Fatalf("unexpected object id length: %d %d", len(first), len(second))
	}
	if first == second {
		t.Fatal("expected random object ids to differ")
	}
}

func TestAvatarObjectName(t *testing.T) {
	storage := &AvatarStorage{bucket: "avatars"}

	objectName := storage.avatarObjectName("http://example.com/avatars/users/7/avatar.jpg")
	if objectName != "users/7/avatar.jpg" {
		t.Fatalf("unexpected object name: %s", objectName)
	}

	objectName = storage.avatarObjectName("http://example.com/files/users/7/avatar.jpg")
	if objectName != "" {
		t.Fatalf("expected empty object name, got %s", objectName)
	}
}
