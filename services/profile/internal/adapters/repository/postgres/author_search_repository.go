package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/usecase"
	"github.com/lib/pq"
	"strings"
)

func (repository *ProfileRepository) SearchAuthors(ctx context.Context, query usecase.SearchAuthorsQuery) ([]domain.AuthorSummary, error) {
	const baseQuery = `
		WITH filtered_authors AS (
			SELECT p.user_id
			FROM profile p
			WHERE p.is_trainer = TRUE
	`
	const selectQuery = `
		)
		SELECT
			p.user_id,
			p.username,
			p.first_name,
			p.last_name,
			p.bio,
			p.avatar_url,
			tp.education_degree,
			tp.career_since_date,
			ts.sport_type_id,
			ts.experience_years,
			ts.sports_rank
		FROM filtered_authors fa
		JOIN profile p ON p.user_id = fa.user_id
		LEFT JOIN trainer_profile tp ON tp.user_id = p.user_id
		LEFT JOIN trainer_sport ts ON ts.user_id = p.user_id
	`

	args := []any{}
	conditions := make([]string, 0, 2)

	if trimmedQuery := strings.TrimSpace(query.Query); trimmedQuery != "" {
		args = append(args, "%"+trimmedQuery+"%")
		placeholder := fmt.Sprintf("$%d", len(args))
		conditions = append(conditions, "(p.username ILIKE "+placeholder+" OR p.first_name ILIKE "+placeholder+" OR p.last_name ILIKE "+placeholder+" OR p.bio ILIKE "+placeholder+")")
	}
	if len(query.SportTypeIDs) > 0 || query.MinExperienceYears != nil || query.MaxExperienceYears != nil || query.OnlyWithRank {
		sportConditions := []string{"filter_ts.user_id = p.user_id"}
		if len(query.SportTypeIDs) > 0 {
			args = append(args, pq.Array(query.SportTypeIDs))
			placeholder := fmt.Sprintf("$%d", len(args))
			sportConditions = append(sportConditions, "filter_ts.sport_type_id = ANY("+placeholder+")")
		}
		if query.MinExperienceYears != nil {
			args = append(args, *query.MinExperienceYears)
			placeholder := fmt.Sprintf("$%d", len(args))
			sportConditions = append(sportConditions, "filter_ts.experience_years >= "+placeholder)
		}
		if query.MaxExperienceYears != nil {
			args = append(args, *query.MaxExperienceYears)
			placeholder := fmt.Sprintf("$%d", len(args))
			sportConditions = append(sportConditions, "filter_ts.experience_years <= "+placeholder)
		}
		if query.OnlyWithRank {
			sportConditions = append(sportConditions, "filter_ts.sports_rank IS NOT NULL AND btrim(filter_ts.sports_rank) <> ''")
		}

		conditions = append(conditions, "EXISTS (SELECT 1 FROM trainer_sport filter_ts WHERE "+strings.Join(sportConditions, " AND ")+")")
	}

	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(baseQuery)
	for _, condition := range conditions {
		queryBuilder.WriteString(" AND ")
		queryBuilder.WriteString(condition)
	}

	args = append(args, query.Limit, query.Offset)
	queryBuilder.WriteString(fmt.Sprintf(" ORDER BY p.user_id DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args)))
	queryBuilder.WriteString(selectQuery)
	queryBuilder.WriteString(" ORDER BY p.user_id DESC, ts.sport_type_id")

	rows, err := repository.db.QueryContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	authors := make([]domain.AuthorSummary, 0)
	indexByID := make(map[int64]int)
	for rows.Next() {
		var (
			author          domain.AuthorSummary
			bio             sql.NullString
			avatarURL       sql.NullString
			educationDegree sql.NullString
			careerSinceDate sql.NullTime
			sportTypeID     sql.NullInt64
			experienceYears sql.NullInt64
			sportsRank      sql.NullString
		)

		if err := rows.Scan(
			&author.UserID,
			&author.Username,
			&author.FirstName,
			&author.LastName,
			&bio,
			&avatarURL,
			&educationDegree,
			&careerSinceDate,
			&sportTypeID,
			&experienceYears,
			&sportsRank,
		); err != nil {
			return nil, err
		}

		index, ok := indexByID[author.UserID]
		if !ok {
			if bio.Valid {
				author.Bio = &bio.String
			}
			if avatarURL.Valid {
				author.AvatarURL = &avatarURL.String
			}
			author.TrainerDetails = &domain.TrainerDetails{
				Sports: make([]domain.TrainerSport, 0),
			}
			if educationDegree.Valid {
				author.TrainerDetails.EducationDegree = &educationDegree.String
			}
			if careerSinceDate.Valid {
				author.TrainerDetails.CareerSinceDate = &careerSinceDate.Time
			}

			authors = append(authors, author)
			index = len(authors) - 1
			indexByID[author.UserID] = index
		}

		if sportTypeID.Valid {
			sport := domain.TrainerSport{
				SportTypeID:     sportTypeID.Int64,
				ExperienceYears: int(experienceYears.Int64),
			}
			if sportsRank.Valid {
				sport.SportsRank = &sportsRank.String
			}
			authors[index].TrainerDetails.Sports = append(authors[index].TrainerDetails.Sports, sport)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return authors, nil
}
