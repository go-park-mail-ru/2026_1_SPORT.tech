package handler

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/usecase"
)

type sportTypeUseCaseStub struct {
	listSportTypesFunc func(ctx context.Context) ([]domain.SportType, error)
}

func (stub *sportTypeUseCaseStub) ListSportTypes(ctx context.Context) ([]domain.SportType, error) {
	return stub.listSportTypesFunc(ctx)
}

type sessionUseCaseStub struct {
	createSessionFunc        func(ctx context.Context, userID int64) (string, error)
	getUserIDBySessionIDFunc func(ctx context.Context, sessionID string) (int64, error)
	revokeSessionFunc        func(ctx context.Context, sessionID string) error
}

func (stub *sessionUseCaseStub) CreateSession(ctx context.Context, userID int64) (string, error) {
	return stub.createSessionFunc(ctx, userID)
}

func (stub *sessionUseCaseStub) GetUserIDBySessionID(ctx context.Context, sessionID string) (int64, error) {
	return stub.getUserIDBySessionIDFunc(ctx, sessionID)
}

func (stub *sessionUseCaseStub) RevokeSession(ctx context.Context, sessionID string) error {
	return stub.revokeSessionFunc(ctx, sessionID)
}

type userUseCaseStub struct {
	listTrainersFunc    func(ctx context.Context) ([]domain.TrainerListItem, error)
	getByIDFunc         func(ctx context.Context, userID int64) (domain.User, error)
	registerClientFunc  func(ctx context.Context, command usecase.RegisterClientCommand) (domain.User, error)
	registerTrainerFunc func(ctx context.Context, command usecase.RegisterTrainerCommand) (domain.User, error)
	authenticateFunc    func(ctx context.Context, email string, password string) (domain.User, error)
	updateProfileFunc   func(ctx context.Context, userID int64, command usecase.UpdateProfileCommand) (domain.User, error)
	uploadAvatarFunc    func(ctx context.Context, userID int64, fileName string, contentType string, file io.Reader, size int64) (domain.User, error)
	deleteAvatarFunc    func(ctx context.Context, userID int64) error
}

func (stub *userUseCaseStub) ListTrainers(ctx context.Context) ([]domain.TrainerListItem, error) {
	return stub.listTrainersFunc(ctx)
}
func (stub *userUseCaseStub) GetByID(ctx context.Context, userID int64) (domain.User, error) {
	return stub.getByIDFunc(ctx, userID)
}
func (stub *userUseCaseStub) RegisterClient(ctx context.Context, command usecase.RegisterClientCommand) (domain.User, error) {
	return stub.registerClientFunc(ctx, command)
}
func (stub *userUseCaseStub) RegisterTrainer(ctx context.Context, command usecase.RegisterTrainerCommand) (domain.User, error) {
	return stub.registerTrainerFunc(ctx, command)
}
func (stub *userUseCaseStub) Authenticate(ctx context.Context, email string, password string) (domain.User, error) {
	return stub.authenticateFunc(ctx, email, password)
}
func (stub *userUseCaseStub) UpdateProfile(ctx context.Context, userID int64, command usecase.UpdateProfileCommand) (domain.User, error) {
	return stub.updateProfileFunc(ctx, userID, command)
}
func (stub *userUseCaseStub) UploadAvatar(ctx context.Context, userID int64, fileName string, contentType string, file io.Reader, size int64) (domain.User, error) {
	return stub.uploadAvatarFunc(ctx, userID, fileName, contentType, file, size)
}
func (stub *userUseCaseStub) DeleteAvatar(ctx context.Context, userID int64) error {
	return stub.deleteAvatarFunc(ctx, userID)
}

type postUseCaseStub struct {
	listProfilePostsFunc func(ctx context.Context, profileUserID int64, currentUserID int64) ([]domain.PostListItem, error)
	getByIDFunc          func(ctx context.Context, postID int64, currentUserID int64) (domain.Post, error)
	setLikeFunc          func(ctx context.Context, postID int64, userID int64) (domain.PostLikeStatus, error)
	deleteLikeFunc       func(ctx context.Context, postID int64, userID int64) (domain.PostLikeStatus, error)
	createFunc           func(ctx context.Context, trainerID int64, command usecase.CreatePostCommand) (domain.Post, error)
	updateFunc           func(ctx context.Context, trainerID int64, postID int64, command usecase.UpdatePostCommand) (domain.Post, error)
	deleteFunc           func(ctx context.Context, trainerID int64, postID int64) error
}

func (stub *postUseCaseStub) ListProfilePosts(ctx context.Context, profileUserID int64, currentUserID int64) ([]domain.PostListItem, error) {
	return stub.listProfilePostsFunc(ctx, profileUserID, currentUserID)
}
func (stub *postUseCaseStub) GetByID(ctx context.Context, postID int64, currentUserID int64) (domain.Post, error) {
	return stub.getByIDFunc(ctx, postID, currentUserID)
}
func (stub *postUseCaseStub) SetLike(ctx context.Context, postID int64, userID int64) (domain.PostLikeStatus, error) {
	return stub.setLikeFunc(ctx, postID, userID)
}
func (stub *postUseCaseStub) DeleteLike(ctx context.Context, postID int64, userID int64) (domain.PostLikeStatus, error) {
	return stub.deleteLikeFunc(ctx, postID, userID)
}
func (stub *postUseCaseStub) Create(ctx context.Context, trainerID int64, command usecase.CreatePostCommand) (domain.Post, error) {
	return stub.createFunc(ctx, trainerID, command)
}
func (stub *postUseCaseStub) Update(ctx context.Context, trainerID int64, postID int64, command usecase.UpdatePostCommand) (domain.Post, error) {
	return stub.updateFunc(ctx, trainerID, postID, command)
}
func (stub *postUseCaseStub) Delete(ctx context.Context, trainerID int64, postID int64) error {
	return stub.deleteFunc(ctx, trainerID, postID)
}

type donationUseCaseStub struct {
	createFunc func(ctx context.Context, command usecase.CreateDonationCommand) (domain.Donation, error)
}

func (stub *donationUseCaseStub) Create(ctx context.Context, command usecase.CreateDonationCommand) (domain.Donation, error) {
	return stub.createFunc(ctx, command)
}

func TestDocsAndHealthHandlers(t *testing.T) {
	handler := &Handler{}

	recorder := httptest.NewRecorder()
	handler.handleHealth(recorder, httptest.NewRequest(http.MethodGet, "/health", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected health status: %d", recorder.Code)
	}
}

func TestSportTypeAndDonationHandlers(t *testing.T) {
	t.Run("sport types success", func(t *testing.T) {
		handler := &Handler{
			sportTypeUseCase: &sportTypeUseCaseStub{
				listSportTypesFunc: func(ctx context.Context) ([]domain.SportType, error) {
					return []domain.SportType{{ID: 1, Name: "Бег"}}, nil
				},
			},
		}

		recorder := httptest.NewRecorder()
		handler.handleGetSportTypes(recorder, httptest.NewRequest(http.MethodGet, "/sport-types", nil))
		if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "sport_type_id") {
			t.Fatalf("unexpected response: %d %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("donation success", func(t *testing.T) {
		handler := &Handler{
			donationUseCase: &donationUseCaseStub{
				createFunc: func(ctx context.Context, command usecase.CreateDonationCommand) (domain.Donation, error) {
					return domain.Donation{DonationID: 11, SenderUserID: 1, RecipientUserID: 3, AmountValue: 5000, Currency: "RUB", CreatedAt: time.Now()}, nil
				},
			},
		}

		request := httptest.NewRequest(http.MethodPost, "/profiles/3/donations", strings.NewReader(`{"amount_value":5000,"currency":"RUB"}`))
		request.SetPathValue("user_id", "3")
		request = request.WithContext(context.WithValue(request.Context(), userIDContextKey, int64(1)))
		recorder := httptest.NewRecorder()

		handler.handlePostProfileDonation(recorder, request)
		if recorder.Code != http.StatusCreated {
			t.Fatalf("unexpected response: %d %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("donation recipient not found", func(t *testing.T) {
		handler := &Handler{
			donationUseCase: &donationUseCaseStub{
				createFunc: func(ctx context.Context, command usecase.CreateDonationCommand) (domain.Donation, error) {
					return domain.Donation{}, usecase.ErrDonationRecipientNotFound
				},
			},
		}

		request := httptest.NewRequest(http.MethodPost, "/profiles/3/donations", strings.NewReader(`{"amount_value":5000,"currency":"RUB"}`))
		request.SetPathValue("user_id", "3")
		request = request.WithContext(context.WithValue(request.Context(), userIDContextKey, int64(1)))
		recorder := httptest.NewRecorder()

		handler.handlePostProfileDonation(recorder, request)
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("unexpected response: %d %s", recorder.Code, recorder.Body.String())
		}
	})
}

func TestProfileHelpers(t *testing.T) {
	request := httptest.NewRequest(http.MethodPatch, "/profiles/me", strings.NewReader(`{"username":"john_doe","bio":null}`))
	command, validationErrors, err := decodeUpdateProfileRequest(request)
	if err != nil || len(validationErrors) != 0 {
		t.Fatalf("unexpected decode result: %v %+v", err, validationErrors)
	}
	if !command.HasUsername || command.Username != "john_doe" || !command.HasBio || command.Bio != nil {
		t.Fatalf("unexpected command: %+v", command)
	}

	request = httptest.NewRequest(http.MethodPatch, "/profiles/me", strings.NewReader(`{"unknown":"x"}`))
	if _, _, err := decodeUpdateProfileRequest(request); err == nil {
		t.Fatal("expected unknown field error")
	}

	request = httptest.NewRequest(http.MethodPatch, "/profiles/me", strings.NewReader(`{"trainer_details":{"education_degree":"Bachelor","career_since_date":"2020-01-01","sports":[{"sport_type_id":1,"experience_years":3}]}}`))
	command, validationErrors, err = decodeUpdateProfileRequest(request)
	if err != nil || len(validationErrors) != 0 {
		t.Fatalf("unexpected trainer details decode result: %v %+v", err, validationErrors)
	}
	if !command.HasEducationDegree || command.EducationDegree == nil || *command.EducationDegree != "Bachelor" || !command.HasCareerSinceDate || !command.HasSports || len(command.Sports) != 1 {
		t.Fatalf("unexpected trainer command: %+v", command)
	}

	content, contentType, validationErrors, err := decodeAvatarFile(bytes.NewReader(minimalPNG()))
	if err != nil || len(validationErrors) != 0 || contentType != "image/png" || len(content) == 0 {
		t.Fatalf("unexpected avatar decode result: %v %+v %q", err, validationErrors, contentType)
	}

	_, _, validationErrors, err = decodeAvatarFile(bytes.NewReader(nil))
	if err != nil || len(validationErrors) == 0 {
		t.Fatalf("expected empty file validation error: %v %+v", err, validationErrors)
	}

	sessionUseCase := &sessionUseCaseStub{
		getUserIDBySessionIDFunc: func(ctx context.Context, sessionID string) (int64, error) {
			return 7, nil
		},
	}
	handler := &Handler{sessionUseCase: sessionUseCase, authCookieName: "sid"}
	request = httptest.NewRequest(http.MethodGet, "/profiles/7", nil)
	request.AddCookie(&http.Cookie{Name: "sid", Value: "cookie"})
	isMe, err := handler.isCurrentUser(request, 7)
	if err != nil || !isMe {
		t.Fatalf("unexpected isMe result: %v %v", isMe, err)
	}

	profile := handler.newProfileResponse(domain.User{
		ID:        7,
		Username:  "coach",
		FirstName: "John",
		LastName:  "Doe",
		AvatarURL: stringPtr("http://localhost:8000/avatars/users/7/avatar.jpg"),
		TrainerDetails: &domain.TrainerDetails{
			CareerSinceDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			Sports:          []domain.TrainerSport{{SportTypeID: 1, ExperienceYears: 5}},
		},
	}, true)
	if profile.UserID != 7 || profile.TrainerDetails == nil || len(profile.TrainerDetails.Sports) != 1 {
		t.Fatalf("unexpected profile response: %+v", profile)
	}
}

func TestPostHelpersAndLikeHandlers(t *testing.T) {
	request := httptest.NewRequest(http.MethodPatch, "/posts/1", strings.NewReader(`{"title":"new title","attachments":[{"kind":"image","file_url":"http://example.com/img.jpg"}]}`))
	command, validationErrors, err := decodeUpdatePostRequest(request)
	if err != nil || len(validationErrors) != 0 {
		t.Fatalf("unexpected decode result: %v %+v", err, validationErrors)
	}
	if command.Title == nil || *command.Title != "new title" || !command.HasAttachments || len(command.Attachments) != 1 {
		t.Fatalf("unexpected command: %+v", command)
	}

	request = httptest.NewRequest(http.MethodPatch, "/posts/1", strings.NewReader(`{"attachments":null}`))
	if _, _, err := decodeUpdatePostRequest(request); err == nil {
		t.Fatal("expected attachments null error")
	}

	validationErrors = validateCreatePostRequest(createPostRequest{
		Title:       "",
		TextContent: "",
		Attachments: []createPostAttachmentRequest{{Kind: "bad", FileURL: ""}},
	})
	if len(validationErrors) < 3 {
		t.Fatalf("expected validation errors, got %+v", validationErrors)
	}

	request = httptest.NewRequest(http.MethodGet, "/posts/15", nil)
	request.SetPathValue("post_id", "15")
	postID, ok := parsePostID(request)
	if !ok || postID != 15 {
		t.Fatalf("unexpected post id: %d %v", postID, ok)
	}

	response := newPostResponse(domain.Post{
		PostID:      15,
		TrainerID:   7,
		Title:       "Title",
		TextContent: "Content",
		LikesCount:  3,
		IsLiked:     true,
		Attachments: []domain.PostAttachment{{PostAttachmentID: 1, Kind: "image", FileURL: "http://example.com/img.jpg"}},
	})
	if response.PostID != 15 || len(response.Attachments) != 1 || response.LikesCount != 3 || !response.IsLiked {
		t.Fatalf("unexpected post response: %+v", response)
	}

	likeResponse := newPostLikeResponse(domain.PostLikeStatus{PostID: 15, LikesCount: 3, IsLiked: true})
	if likeResponse.PostID != 15 || !likeResponse.IsLiked {
		t.Fatalf("unexpected like response: %+v", likeResponse)
	}

	handler := &Handler{
		postUseCase: &postUseCaseStub{
			setLikeFunc: func(ctx context.Context, postID int64, userID int64) (domain.PostLikeStatus, error) {
				return domain.PostLikeStatus{PostID: postID, LikesCount: 2, IsLiked: true}, nil
			},
			deleteLikeFunc: func(ctx context.Context, postID int64, userID int64) (domain.PostLikeStatus, error) {
				return domain.PostLikeStatus{PostID: postID, LikesCount: 1, IsLiked: false}, nil
			},
		},
	}

	request = httptest.NewRequest(http.MethodPost, "/posts/15/likes", nil)
	request.SetPathValue("post_id", "15")
	request = request.WithContext(context.WithValue(request.Context(), userIDContextKey, int64(7)))
	recorder := httptest.NewRecorder()
	handler.handlePostPostLike(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected like status: %d %s", recorder.Code, recorder.Body.String())
	}

	request = httptest.NewRequest(http.MethodDelete, "/posts/15/likes", nil)
	request.SetPathValue("post_id", "15")
	request = request.WithContext(context.WithValue(request.Context(), userIDContextKey, int64(7)))
	recorder = httptest.NewRecorder()
	handler.handleDeletePostLike(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected unlike status: %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestAuthValidationHelpers(t *testing.T) {
	if sessionCookieSameSite() != http.SameSiteLaxMode {
		t.Fatalf("unexpected same site mode: %v", sessionCookieSameSite())
	}

	errors := validateClientRegisterRequest(clientRegisterRequest{
		Username:       "x",
		Email:          "bad",
		Password:       "short",
		PasswordRepeat: "other",
		FirstName:      "",
		LastName:       "",
	})
	if len(errors) < 5 {
		t.Fatalf("expected many validation errors, got %+v", errors)
	}

	degree := strings.Repeat("a", 256)
	rank := strings.Repeat("b", 101)
	_, errors = validateTrainerRegisterRequest(trainerRegisterRequest{
		Username:       "trainer",
		Email:          "trainer@example.com",
		Password:       "supersecret123",
		PasswordRepeat: "supersecret123",
		FirstName:      "John",
		LastName:       "Doe",
		TrainerDetails: trainerRegisterDetailsRequest{
			EducationDegree: &degree,
			CareerSinceDate: "3020-01-01",
			Sports: []trainerRegisterSportRequest{
				{SportTypeID: 1, ExperienceYears: -1, SportsRank: &rank},
				{SportTypeID: 1, ExperienceYears: 1},
			},
		},
	})
	if len(errors) == 0 {
		t.Fatal("expected trainer validation errors")
	}

	handler := &Handler{storagePublicBaseURL: "http://example.com/avatars"}
	response := handler.newAuthResponse(domain.User{
		ID:        1,
		Username:  "john",
		Email:     "john@example.com",
		AvatarURL: stringPtr("http://localhost:8000/avatars/users/1/avatar.jpg"),
	})
	if response.User.AvatarURL == nil || !strings.Contains(*response.User.AvatarURL, "example.com") {
		t.Fatalf("unexpected auth response: %+v", response)
	}
}

func TestAuthHandlers(t *testing.T) {
	user := domain.User{ID: 5, Username: "john", Email: "john@example.com"}

	t.Run("register client success", func(t *testing.T) {
		handler := &Handler{
			authCookieName: "sid",
			userUseCase: &userUseCaseStub{
				registerClientFunc: func(ctx context.Context, command usecase.RegisterClientCommand) (domain.User, error) {
					return user, nil
				},
			},
			sessionUseCase: &sessionUseCaseStub{
				createSessionFunc: func(ctx context.Context, userID int64) (string, error) { return "session-id", nil },
			},
		}

		request := httptest.NewRequest(http.MethodPost, "/auth/register/client", strings.NewReader(`{"username":"john_doe","email":"john@example.com","password":"supersecret123","password_repeat":"supersecret123","first_name":"John","last_name":"Doe"}`))
		recorder := httptest.NewRecorder()
		handler.handlePostAuthRegisterClient(recorder, request)
		if recorder.Code != http.StatusCreated {
			t.Fatalf("unexpected response: %d %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("register trainer success", func(t *testing.T) {
		handler := &Handler{
			authCookieName: "sid",
			userUseCase: &userUseCaseStub{
				registerTrainerFunc: func(ctx context.Context, command usecase.RegisterTrainerCommand) (domain.User, error) {
					return domain.User{ID: 6, Username: "trainer", Email: "trainer@example.com", IsTrainer: true}, nil
				},
			},
			sessionUseCase: &sessionUseCaseStub{
				createSessionFunc: func(ctx context.Context, userID int64) (string, error) { return "session-id", nil },
			},
		}

		request := httptest.NewRequest(http.MethodPost, "/auth/register/trainer", strings.NewReader(`{
			"username":"trainer",
			"email":"trainer@example.com",
			"password":"supersecret123",
			"password_repeat":"supersecret123",
			"first_name":"John",
			"last_name":"Doe",
			"trainer_details":{
				"career_since_date":"2020-01-01",
				"sports":[{"sport_type_id":1,"experience_years":3}]
			}
		}`))
		recorder := httptest.NewRecorder()
		handler.handlePostAuthRegisterTrainer(recorder, request)
		if recorder.Code != http.StatusCreated {
			t.Fatalf("unexpected response: %d %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("login success", func(t *testing.T) {
		handler := &Handler{
			authCookieName: "sid",
			userUseCase: &userUseCaseStub{
				authenticateFunc: func(ctx context.Context, email string, password string) (domain.User, error) { return user, nil },
			},
			sessionUseCase: &sessionUseCaseStub{
				createSessionFunc: func(ctx context.Context, userID int64) (string, error) { return "session-id", nil },
			},
		}

		request := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"email":"john@example.com","password":"supersecret123"}`))
		recorder := httptest.NewRecorder()
		handler.handlePostAuthLogin(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("unexpected response: %d %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("auth me success", func(t *testing.T) {
		handler := &Handler{
			userUseCase: &userUseCaseStub{
				getByIDFunc: func(ctx context.Context, userID int64) (domain.User, error) { return user, nil },
			},
		}
		request := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
		request = request.WithContext(context.WithValue(request.Context(), userIDContextKey, int64(5)))
		recorder := httptest.NewRecorder()
		handler.handleGetAuthMe(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("unexpected response: %d %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("logout success", func(t *testing.T) {
		handler := &Handler{
			authCookieName: "sid",
			sessionUseCase: &sessionUseCaseStub{
				revokeSessionFunc: func(ctx context.Context, sessionID string) error { return nil },
			},
		}
		request := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
		request.AddCookie(&http.Cookie{Name: "sid", Value: "session-id"})
		recorder := httptest.NewRecorder()
		handler.handlePostAuthLogout(recorder, request)
		if recorder.Code != http.StatusNoContent {
			t.Fatalf("unexpected response: %d %s", recorder.Code, recorder.Body.String())
		}
	})
}

func TestProfileAndPostHandlers(t *testing.T) {
	now := time.Now()

	t.Run("get trainers success", func(t *testing.T) {
		handler := &Handler{
			storagePublicBaseURL: "http://example.com/avatars",
			userUseCase: &userUseCaseStub{
				listTrainersFunc: func(ctx context.Context) ([]domain.TrainerListItem, error) {
					return []domain.TrainerListItem{{
						ID:        7,
						Username:  "coach",
						FirstName: "John",
						LastName:  "Doe",
						Bio:       stringPtr("Тренер по бегу"),
						AvatarURL: stringPtr("http://localhost:8000/avatars/users/7/avatar.jpg"),
						TrainerDetails: &domain.TrainerDetails{
							CareerSinceDate: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
							Sports: []domain.TrainerSport{{
								SportTypeID:     1,
								ExperienceYears: 5,
							}},
						},
					}}, nil
				},
			},
		}

		recorder := httptest.NewRecorder()
		handler.handleGetTrainers(recorder, httptest.NewRequest(http.MethodGet, "/trainers", nil))
		if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "\"trainers\"") || !strings.Contains(recorder.Body.String(), "example.com") {
			t.Fatalf("unexpected response: %d %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("get profile success", func(t *testing.T) {
		handler := &Handler{
			authCookieName: "sid",
			userUseCase: &userUseCaseStub{
				getByIDFunc: func(ctx context.Context, userID int64) (domain.User, error) {
					return domain.User{ID: userID, Username: "coach", FirstName: "John", LastName: "Doe"}, nil
				},
			},
			sessionUseCase: &sessionUseCaseStub{
				getUserIDBySessionIDFunc: func(ctx context.Context, sessionID string) (int64, error) { return 7, nil },
			},
		}
		request := httptest.NewRequest(http.MethodGet, "/profiles/7", nil)
		request.SetPathValue("user_id", "7")
		request.AddCookie(&http.Cookie{Name: "sid", Value: "cookie"})
		recorder := httptest.NewRecorder()
		handler.handleGetProfile(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("unexpected response: %d %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("get profile posts success", func(t *testing.T) {
		handler := &Handler{
			userUseCase: &userUseCaseStub{
				getByIDFunc: func(ctx context.Context, userID int64) (domain.User, error) {
					return domain.User{ID: userID}, nil
				},
			},
			postUseCase: &postUseCaseStub{
				listProfilePostsFunc: func(ctx context.Context, profileUserID int64, currentUserID int64) ([]domain.PostListItem, error) {
					return []domain.PostListItem{{PostID: 1, TrainerID: profileUserID, Title: "post", CreatedAt: now, CanView: true, LikesCount: 5, IsLiked: true}}, nil
				},
			},
			sessionUseCase: &sessionUseCaseStub{
				getUserIDBySessionIDFunc: func(ctx context.Context, sessionID string) (int64, error) { return 7, nil },
			},
			authCookieName: "sid",
		}
		request := httptest.NewRequest(http.MethodGet, "/profiles/7/posts", nil)
		request.SetPathValue("user_id", "7")
		request.AddCookie(&http.Cookie{Name: "sid", Value: "cookie"})
		recorder := httptest.NewRecorder()
		handler.handleGetProfilePosts(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("unexpected response: %d %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("get post success", func(t *testing.T) {
		handler := &Handler{
			postUseCase: &postUseCaseStub{
				getByIDFunc: func(ctx context.Context, postID int64, currentUserID int64) (domain.Post, error) {
					return domain.Post{PostID: postID, TrainerID: 7, Title: "post", TextContent: "content", LikesCount: 2, IsLiked: true}, nil
				},
			},
			sessionUseCase: &sessionUseCaseStub{
				getUserIDBySessionIDFunc: func(ctx context.Context, sessionID string) (int64, error) { return 7, nil },
			},
			authCookieName: "sid",
		}
		request := httptest.NewRequest(http.MethodGet, "/posts/5", nil)
		request.SetPathValue("post_id", "5")
		request.AddCookie(&http.Cookie{Name: "sid", Value: "cookie"})
		recorder := httptest.NewRecorder()
		handler.handleGetPost(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("unexpected response: %d %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("create patch delete post success", func(t *testing.T) {
		handler := &Handler{
			userUseCase: &userUseCaseStub{
				getByIDFunc: func(ctx context.Context, userID int64) (domain.User, error) {
					return domain.User{ID: userID, IsTrainer: true}, nil
				},
			},
			postUseCase: &postUseCaseStub{
				createFunc: func(ctx context.Context, trainerID int64, command usecase.CreatePostCommand) (domain.Post, error) {
					return domain.Post{PostID: 1, TrainerID: trainerID, Title: command.Title, TextContent: command.TextContent, LikesCount: 0, IsLiked: false}, nil
				},
				updateFunc: func(ctx context.Context, trainerID int64, postID int64, command usecase.UpdatePostCommand) (domain.Post, error) {
					title := ""
					if command.Title != nil {
						title = *command.Title
					}
					return domain.Post{PostID: postID, TrainerID: trainerID, Title: title, LikesCount: 0, IsLiked: false}, nil
				},
				deleteFunc: func(ctx context.Context, trainerID int64, postID int64) error { return nil },
			},
		}

		createRequest := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(`{"title":"title","text_content":"content","attachments":[]}`))
		createRequest = createRequest.WithContext(context.WithValue(createRequest.Context(), userIDContextKey, int64(7)))
		recorder := httptest.NewRecorder()
		handler.handlePostCreate(recorder, createRequest)
		if recorder.Code != http.StatusCreated {
			t.Fatalf("unexpected create response: %d %s", recorder.Code, recorder.Body.String())
		}

		patchRequest := httptest.NewRequest(http.MethodPatch, "/posts/1", strings.NewReader(`{"title":"updated"}`))
		patchRequest.SetPathValue("post_id", "1")
		patchRequest = patchRequest.WithContext(context.WithValue(patchRequest.Context(), userIDContextKey, int64(7)))
		recorder = httptest.NewRecorder()
		handler.handlePatchPost(recorder, patchRequest)
		if recorder.Code != http.StatusOK {
			t.Fatalf("unexpected patch response: %d %s", recorder.Code, recorder.Body.String())
		}

		deleteRequest := httptest.NewRequest(http.MethodDelete, "/posts/1", nil)
		deleteRequest.SetPathValue("post_id", "1")
		deleteRequest = deleteRequest.WithContext(context.WithValue(deleteRequest.Context(), userIDContextKey, int64(7)))
		recorder = httptest.NewRecorder()
		handler.handleDeletePost(recorder, deleteRequest)
		if recorder.Code != http.StatusNoContent {
			t.Fatalf("unexpected delete response: %d %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("patch profile me success", func(t *testing.T) {
		handler := &Handler{
			userUseCase: &userUseCaseStub{
				updateProfileFunc: func(ctx context.Context, userID int64, command usecase.UpdateProfileCommand) (domain.User, error) {
					return domain.User{
						ID:        userID,
						Username:  command.Username,
						FirstName: "John",
						LastName:  "Doe",
						IsTrainer: true,
						TrainerDetails: &domain.TrainerDetails{
							EducationDegree: command.EducationDegree,
							CareerSinceDate: command.CareerSinceDate,
							Sports:          []domain.TrainerSport{{SportTypeID: 1, ExperienceYears: 3}},
						},
					}, nil
				},
			},
		}

		request := httptest.NewRequest(http.MethodPatch, "/profiles/me", strings.NewReader(`{"username":"updated_user","trainer_details":{"education_degree":"Bachelor","career_since_date":"2020-01-01","sports":[{"sport_type_id":1,"experience_years":3}]}}`))
		request = request.WithContext(context.WithValue(request.Context(), userIDContextKey, int64(7)))
		recorder := httptest.NewRecorder()
		handler.handlePatchProfileMe(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("unexpected response: %d %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("upload avatar success", func(t *testing.T) {
		handler := &Handler{
			userUseCase: &userUseCaseStub{
				uploadAvatarFunc: func(ctx context.Context, userID int64, fileName string, contentType string, file io.Reader, size int64) (domain.User, error) {
					content, err := io.ReadAll(file)
					if err != nil {
						t.Fatalf("read upload body: %v", err)
					}
					if userID != 7 || fileName != "avatar.png" || contentType != "image/png" || size != int64(len(content)) || len(content) == 0 {
						t.Fatalf("unexpected upload args: userID=%d fileName=%s contentType=%s size=%d len=%d", userID, fileName, contentType, size, len(content))
					}

					return domain.User{
						ID:        userID,
						AvatarURL: stringPtr("http://cdn.example.com/avatar.png"),
					}, nil
				},
			},
		}

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("avatar", "avatar.png")
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		if _, err := part.Write(minimalPNG()); err != nil {
			t.Fatalf("write png: %v", err)
		}
		if err := writer.Close(); err != nil {
			t.Fatalf("close multipart writer: %v", err)
		}

		request := httptest.NewRequest(http.MethodPost, "/profiles/me/avatar", body)
		request.Header.Set("Content-Type", writer.FormDataContentType())
		request = request.WithContext(context.WithValue(request.Context(), userIDContextKey, int64(7)))
		recorder := httptest.NewRecorder()
		handler.handlePostProfileAvatar(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("unexpected response: %d %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("delete avatar success", func(t *testing.T) {
		handler := &Handler{
			userUseCase: &userUseCaseStub{
				deleteAvatarFunc: func(ctx context.Context, userID int64) error {
					if userID != 7 {
						t.Fatalf("unexpected user id: %d", userID)
					}
					return nil
				},
			},
		}

		request := httptest.NewRequest(http.MethodDelete, "/profiles/me/avatar", nil)
		request = request.WithContext(context.WithValue(request.Context(), userIDContextKey, int64(7)))
		recorder := httptest.NewRecorder()
		handler.handleDeleteProfileAvatar(recorder, request)
		if recorder.Code != http.StatusNoContent {
			t.Fatalf("unexpected response: %d %s", recorder.Code, recorder.Body.String())
		}
	})
}

func minimalPNG() []byte {
	return []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
		0x54, 0x08, 0xD7, 0x63, 0xF8, 0xCF, 0xC0, 0x00,
		0x00, 0x03, 0x01, 0x01, 0x00, 0x18, 0xDD, 0x8D,
		0x18, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E,
		0x44, 0xAE, 0x42, 0x60, 0x82,
	}
}

func stringPtr(value string) *string {
	return &value
}
