package about

import "time"

// Experience is a structured career entry for the personal about page.
type Experience struct {
	ID           uint64                  `gorm:"primaryKey" json:"id"`
	Company      string                  `gorm:"size:160;not null" json:"company"`
	Location     string                  `gorm:"size:160" json:"location"`
	StartDate    time.Time               `json:"startDate"`
	EndDate      *time.Time              `json:"endDate,omitempty"`
	IsCurrent    bool                    `gorm:"not null;default:false" json:"isCurrent"`
	SortOrder    int32                   `gorm:"not null;default:0" json:"sortOrder"`
	Active       bool                    `gorm:"not null;default:true" json:"active"`
	Technologies []string                `gorm:"type:jsonb;serializer:json;not null;default:'[]'" json:"technologies"`
	Translations []ExperienceTranslation `gorm:"foreignKey:ExperienceID" json:"translations,omitempty"`
	CreatedAt    time.Time               `json:"createdAt"`
	UpdatedAt    time.Time               `json:"updatedAt"`
}

type ExperienceTranslation struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	ExperienceID uint64    `gorm:"not null;constraint:OnDelete:CASCADE" json:"experienceId"`
	Locale       string    `gorm:"size:8;not null" json:"locale"`
	Role         string    `gorm:"size:180;not null" json:"role"`
	Summary      string    `gorm:"type:text" json:"summary"`
	Highlights   []string  `gorm:"type:jsonb;serializer:json;not null;default:'[]'" json:"highlights"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// Education is a structured education/training entry for the personal about page.
type Education struct {
	ID           uint64                 `gorm:"primaryKey" json:"id"`
	Institution  string                 `gorm:"size:180;not null" json:"institution"`
	Location     string                 `gorm:"size:160" json:"location"`
	StartDate    time.Time              `json:"startDate"`
	EndDate      *time.Time             `json:"endDate,omitempty"`
	IsCurrent    bool                   `gorm:"not null;default:false" json:"isCurrent"`
	SortOrder    int32                  `gorm:"not null;default:0" json:"sortOrder"`
	Active       bool                   `gorm:"not null;default:true" json:"active"`
	Translations []EducationTranslation `gorm:"foreignKey:EducationID" json:"translations,omitempty"`
	CreatedAt    time.Time              `json:"createdAt"`
	UpdatedAt    time.Time              `json:"updatedAt"`
}

type EducationTranslation struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	EducationID uint64    `gorm:"not null;constraint:OnDelete:CASCADE" json:"educationId"`
	Locale      string    `gorm:"size:8;not null" json:"locale"`
	Title       string    `gorm:"size:180;not null" json:"title"`
	Summary     string    `gorm:"type:text" json:"summary"`
	Highlights  []string  `gorm:"type:jsonb;serializer:json;not null;default:'[]'" json:"highlights"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Skill groups related capabilities for display on the about page.
type Skill struct {
	ID           uint64             `gorm:"primaryKey" json:"id"`
	Key          string             `gorm:"size:120;not null;uniqueIndex" json:"key"`
	Category     SkillCategory      `gorm:"size:40;not null" json:"category"`
	Level        SkillLevel         `gorm:"size:40;not null" json:"level"`
	SortOrder    int32              `gorm:"not null;default:0" json:"sortOrder"`
	Active       bool               `gorm:"not null;default:true" json:"active"`
	Translations []SkillTranslation `json:"translations,omitempty"`
	CreatedAt    time.Time          `json:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt"`
}

type SkillTranslation struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	SkillID   uint64    `gorm:"not null;constraint:OnDelete:CASCADE" json:"skillId"`
	Locale    string    `gorm:"size:8;not null" json:"locale"`
	Name      string    `gorm:"size:120;not null" json:"name"`
	Summary   string    `gorm:"type:text" json:"summary"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Interest groups personal interests shown on the about page.
type Interest struct {
	ID           uint64                `gorm:"primaryKey" json:"id"`
	Key          string                `gorm:"size:120;not null;uniqueIndex" json:"key"`
	SortOrder    int32                 `gorm:"not null;default:0" json:"sortOrder"`
	Active       bool                  `gorm:"not null;default:true" json:"active"`
	Translations []InterestTranslation `gorm:"foreignKey:InterestID" json:"translations,omitempty"`
	CreatedAt    time.Time             `json:"createdAt"`
	UpdatedAt    time.Time             `json:"updatedAt"`
}

type InterestTranslation struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	InterestID uint64    `gorm:"not null;constraint:OnDelete:CASCADE" json:"interestId"`
	Locale     string    `gorm:"size:8;not null" json:"locale"`
	Name       string    `gorm:"size:120;not null" json:"name"`
	Summary    string    `gorm:"type:text" json:"summary"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (Experience) TableName() string { return "about_experiences" }

func (ExperienceTranslation) TableName() string { return "about_experience_translations" }

func (Education) TableName() string { return "about_education" }

func (EducationTranslation) TableName() string { return "about_education_translations" }

func (Skill) TableName() string { return "about_skills" }

func (SkillTranslation) TableName() string { return "about_skill_translations" }

func (Interest) TableName() string { return "about_interests" }

func (InterestTranslation) TableName() string { return "about_interest_translations" }

type SkillCategory string

const (
	SkillCategoryBackend        SkillCategory = "backend"
	SkillCategoryFrontend       SkillCategory = "frontend"
	SkillCategoryDevOps         SkillCategory = "devops"
	SkillCategoryTooling        SkillCategory = "tooling"
	SkillCategoryInfrastructure SkillCategory = "infrastructure"
	SkillCategoryOther          SkillCategory = "other"
)

type SkillLevel string

const (
	SkillLevelLearning     SkillLevel = "learning"
	SkillLevelComfortable  SkillLevel = "comfortable"
	SkillLevelStrong       SkillLevel = "strong"
	SkillLevelProfessional SkillLevel = "professional"
)
