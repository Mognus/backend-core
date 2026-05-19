package about

import (
	"context"
	"encoding/json"

	"github.com/Mognus/go-grpc-crud/dbcrud"
	"gorm.io/gorm"
)

var experienceListConfig = dbcrud.ListConfig{
	Preloads:        []string{"Translations"},
	Searchable:      []string{"company", "location"},
	Filterable:      []string{"id", "company", "location", "active", "isCurrent"},
	SortableColumns: []string{"id", "company", "location", "startDate", "endDate", "sortOrder", "active", "createdAt", "updatedAt"},
	DefaultSort:     "sort_order ASC, start_date DESC, id ASC",
}

var educationListConfig = dbcrud.ListConfig{
	Preloads:        []string{"Translations"},
	Searchable:      []string{"institution", "location"},
	Filterable:      []string{"id", "institution", "location", "active", "isCurrent"},
	SortableColumns: []string{"id", "institution", "location", "startDate", "endDate", "sortOrder", "active", "createdAt", "updatedAt"},
	DefaultSort:     "sort_order ASC, start_date DESC, id ASC",
}

var skillListConfig = dbcrud.ListConfig{
	Preloads:        []string{"Translations"},
	Searchable:      []string{"key", "category", "level"},
	Filterable:      []string{"id", "key", "category", "level", "active"},
	SortableColumns: []string{"id", "key", "category", "level", "sortOrder", "active", "createdAt", "updatedAt"},
	DefaultSort:     "sort_order ASC, id ASC",
}

var interestListConfig = dbcrud.ListConfig{
	Preloads:        []string{"Translations"},
	Searchable:      []string{"key"},
	Filterable:      []string{"id", "key", "active"},
	SortableColumns: []string{"id", "key", "sortOrder", "active", "createdAt", "updatedAt"},
	DefaultSort:     "sort_order ASC, id ASC",
}

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) GetExperience(ctx context.Context, id uint64) (Experience, error) {
	return dbcrud.DefaultGet[Experience](ctx, s.db, id, "Translations")
}

func (s *Service) ListExperiences(ctx context.Context, req dbcrud.ListRequest) ([]Experience, int64, error) {
	return dbcrud.DefaultList[Experience](ctx, s.db, req, experienceListConfig)
}

func (s *Service) CreateExperience(ctx context.Context, experience *Experience) (*Experience, error) {
	return dbcrud.DefaultCreate(ctx, s.db, experience, "Translations")
}

func (s *Service) UpdateExperience(ctx context.Context, id uint64, updates map[string]any) (*Experience, error) {
	return dbcrud.DefaultUpdate[Experience](ctx, s.db, id, updates, "Translations")
}

func (s *Service) SaveExperience(ctx context.Context, id uint64, experience *Experience) (*Experience, error) {
	return dbcrud.DefaultReplaceChildren[Experience, ExperienceTranslation](
		ctx,
		s.db,
		id,
		experienceUpdates(experience),
		"experience_id",
		experience.Translations,
		func(translation *ExperienceTranslation, id uint64) {
			translation.ExperienceID = id
		},
		"Translations",
	)
}

func (s *Service) DeleteExperience(ctx context.Context, id uint64) error {
	return dbcrud.DefaultDelete(ctx, s.db, &Experience{}, id)
}

func (s *Service) GetEducation(ctx context.Context, id uint64) (Education, error) {
	return dbcrud.DefaultGet[Education](ctx, s.db, id, "Translations")
}

func (s *Service) ListEducation(ctx context.Context, req dbcrud.ListRequest) ([]Education, int64, error) {
	return dbcrud.DefaultList[Education](ctx, s.db, req, educationListConfig)
}

func (s *Service) CreateEducation(ctx context.Context, education *Education) (*Education, error) {
	return dbcrud.DefaultCreate(ctx, s.db, education, "Translations")
}

func (s *Service) UpdateEducation(ctx context.Context, id uint64, updates map[string]any) (*Education, error) {
	return dbcrud.DefaultUpdate[Education](ctx, s.db, id, updates, "Translations")
}

func (s *Service) SaveEducation(ctx context.Context, id uint64, education *Education) (*Education, error) {
	return dbcrud.DefaultReplaceChildren[Education, EducationTranslation](
		ctx,
		s.db,
		id,
		educationUpdates(education),
		"education_id",
		education.Translations,
		func(translation *EducationTranslation, id uint64) {
			translation.EducationID = id
		},
		"Translations",
	)
}

func (s *Service) DeleteEducation(ctx context.Context, id uint64) error {
	return dbcrud.DefaultDelete(ctx, s.db, &Education{}, id)
}

func (s *Service) GetSkill(ctx context.Context, id uint64) (Skill, error) {
	return dbcrud.DefaultGet[Skill](ctx, s.db, id, "Translations")
}

func (s *Service) ListSkills(ctx context.Context, req dbcrud.ListRequest) ([]Skill, int64, error) {
	return dbcrud.DefaultList[Skill](ctx, s.db, req, skillListConfig)
}

func (s *Service) CreateSkill(ctx context.Context, skill *Skill) (*Skill, error) {
	return dbcrud.DefaultCreate(ctx, s.db, skill, "Translations")
}

func (s *Service) UpdateSkill(ctx context.Context, id uint64, updates map[string]any) (*Skill, error) {
	return dbcrud.DefaultUpdate[Skill](ctx, s.db, id, updates, "Translations")
}

func (s *Service) SaveSkill(ctx context.Context, id uint64, skill *Skill) (*Skill, error) {
	return dbcrud.DefaultReplaceChildren[Skill, SkillTranslation](
		ctx,
		s.db,
		id,
		skillUpdates(skill),
		"skill_id",
		skill.Translations,
		func(translation *SkillTranslation, id uint64) {
			translation.SkillID = id
		},
		"Translations",
	)
}

func (s *Service) DeleteSkill(ctx context.Context, id uint64) error {
	return dbcrud.DefaultDelete(ctx, s.db, &Skill{}, id)
}

func (s *Service) GetInterest(ctx context.Context, id uint64) (Interest, error) {
	return dbcrud.DefaultGet[Interest](ctx, s.db, id, "Translations")
}

func (s *Service) ListInterests(ctx context.Context, req dbcrud.ListRequest) ([]Interest, int64, error) {
	return dbcrud.DefaultList[Interest](ctx, s.db, req, interestListConfig)
}

func (s *Service) CreateInterest(ctx context.Context, interest *Interest) (*Interest, error) {
	return dbcrud.DefaultCreate(ctx, s.db, interest, "Translations")
}

func (s *Service) UpdateInterest(ctx context.Context, id uint64, updates map[string]any) (*Interest, error) {
	return dbcrud.DefaultUpdate[Interest](ctx, s.db, id, updates, "Translations")
}

func (s *Service) SaveInterest(ctx context.Context, id uint64, interest *Interest) (*Interest, error) {
	return dbcrud.DefaultReplaceChildren[Interest, InterestTranslation](
		ctx,
		s.db,
		id,
		interestUpdates(interest),
		"interest_id",
		interest.Translations,
		func(translation *InterestTranslation, id uint64) {
			translation.InterestID = id
		},
		"Translations",
	)
}

func (s *Service) DeleteInterest(ctx context.Context, id uint64) error {
	return dbcrud.DefaultDelete(ctx, s.db, &Interest{}, id)
}

func experienceUpdates(experience *Experience) map[string]any {
	return map[string]any{
		"company":      experience.Company,
		"location":     experience.Location,
		"start_date":   experience.StartDate,
		"end_date":     experience.EndDate,
		"is_current":   experience.IsCurrent,
		"sort_order":   experience.SortOrder,
		"active":       experience.Active,
		"technologies": jsonb(stringSlice(experience.Technologies)),
	}
}

func educationUpdates(education *Education) map[string]any {
	return map[string]any{
		"institution": education.Institution,
		"location":    education.Location,
		"start_date":  education.StartDate,
		"end_date":    education.EndDate,
		"is_current":  education.IsCurrent,
		"sort_order":  education.SortOrder,
		"active":      education.Active,
	}
}

func skillUpdates(skill *Skill) map[string]any {
	return map[string]any{
		"key":        skill.Key,
		"category":   skill.Category,
		"level":      skill.Level,
		"sort_order": skill.SortOrder,
		"active":     skill.Active,
	}
}

func interestUpdates(interest *Interest) map[string]any {
	return map[string]any{
		"key":        interest.Key,
		"sort_order": interest.SortOrder,
		"active":     interest.Active,
	}
}

func jsonb(value any) any {
	body, _ := json.Marshal(value)
	return gorm.Expr("?::jsonb", string(body))
}

func stringSlice(value []string) []string {
	if value == nil {
		return []string{}
	}
	return value
}
