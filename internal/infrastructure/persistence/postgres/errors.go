package postgres

const (
	sqlStateForeignKeyViolation = "23503"
	sqlStateUniqueViolation     = "23505"
)

const (
	postMinTierConstraint                = "post_min_tier_fk"
	userEmailUniqueConstraint            = "user_email_key"
	userProfileUsernameUniqueConstraint  = "user_profile_username_key"
	trainerSportTypeForeignKeyConstraint = "trainer_to_sport_type_sport_type_id_fkey"
)
