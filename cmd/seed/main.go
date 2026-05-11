package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"template/internal/about"
	"template/internal/config"
	"template/internal/platform/db"

	"gorm.io/gorm"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	database, err := db.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}

	switch os.Args[1] {
	case "about", "all":
		seedExperiences(database, experienceSeeds())
		seedEducation(database, educationSeeds())
		seedSkills(database, skillSeeds())
		seedInterests(database, interestSeeds())
	default:
		printUsage()
		os.Exit(1)
	}
}

type experienceSeed struct {
	Company      string
	Location     string
	StartDate    string
	EndDate      string
	IsCurrent    bool
	SortOrder    int32
	Active       bool
	Technologies []string
	Translations []about.ExperienceTranslation
}

type interestSeed struct {
	Key          string
	SortOrder    int32
	Active       bool
	Translations []about.InterestTranslation
}

type educationSeed struct {
	Institution  string
	Location     string
	StartDate    string
	EndDate      string
	IsCurrent    bool
	SortOrder    int32
	Active       bool
	Translations []about.EducationTranslation
}

type skillSeed struct {
	Key          string
	Category     about.SkillCategory
	Level        about.SkillLevel
	SortOrder    int32
	Active       bool
	Translations []about.SkillTranslation
}

func seedExperiences(database *gorm.DB, seeds []experienceSeed) {
	for _, seed := range seeds {
		if err := upsertExperience(database, seed); err != nil {
			log.Fatalf("seed experience %s: %v", seed.Company, err)
		}
		fmt.Printf("seeded experience %s\n", seed.Company)
	}
}

func seedEducation(database *gorm.DB, seeds []educationSeed) {
	for _, seed := range seeds {
		if err := upsertEducation(database, seed); err != nil {
			log.Fatalf("seed education %s: %v", seed.Institution, err)
		}
		fmt.Printf("seeded education %s\n", seed.Institution)
	}
}

func seedSkills(database *gorm.DB, seeds []skillSeed) {
	for _, seed := range seeds {
		if err := upsertSkill(database, seed); err != nil {
			log.Fatalf("seed skill %s: %v", seed.Key, err)
		}
		fmt.Printf("seeded skill %s\n", seed.Key)
	}
}

func seedInterests(database *gorm.DB, seeds []interestSeed) {
	for _, seed := range seeds {
		if err := upsertInterest(database, seed); err != nil {
			log.Fatalf("seed interest %s: %v", seed.Key, err)
		}
		fmt.Printf("seeded interest %s\n", seed.Key)
	}
}

func upsertExperience(database *gorm.DB, seed experienceSeed) error {
	startDate, err := parseDate(seed.StartDate)
	if err != nil {
		return err
	}
	endDate, err := parseOptionalDate(seed.EndDate)
	if err != nil {
		return err
	}

	return database.Transaction(func(tx *gorm.DB) error {
		var item about.Experience
		err := tx.Where("company = ? AND start_date = ?", seed.Company, startDate).First(&item).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			item = about.Experience{Company: seed.Company, StartDate: startDate}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}

		updates := map[string]any{
			"company":      seed.Company,
			"location":     seed.Location,
			"start_date":   startDate,
			"end_date":     endDate,
			"is_current":   seed.IsCurrent,
			"sort_order":   seed.SortOrder,
			"active":       seed.Active,
			"technologies": jsonb(stringSlice(seed.Technologies)),
		}
		if err := tx.Model(&item).Updates(updates).Error; err != nil {
			return err
		}
		return replaceExperienceTranslations(tx, item.ID, seed.Translations)
	})
}

func upsertEducation(database *gorm.DB, seed educationSeed) error {
	startDate, err := parseDate(seed.StartDate)
	if err != nil {
		return err
	}
	endDate, err := parseOptionalDate(seed.EndDate)
	if err != nil {
		return err
	}

	return database.Transaction(func(tx *gorm.DB) error {
		var item about.Education
		err := tx.Where("institution = ? AND start_date = ?", seed.Institution, startDate).First(&item).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			item = about.Education{Institution: seed.Institution, StartDate: startDate}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}

		updates := map[string]any{
			"institution": seed.Institution,
			"location":    seed.Location,
			"start_date":  startDate,
			"end_date":    endDate,
			"is_current":  seed.IsCurrent,
			"sort_order":  seed.SortOrder,
			"active":      seed.Active,
		}
		if err := tx.Model(&item).Updates(updates).Error; err != nil {
			return err
		}
		return replaceEducationTranslations(tx, item.ID, seed.Translations)
	})
}

func upsertSkill(database *gorm.DB, seed skillSeed) error {
	return database.Transaction(func(tx *gorm.DB) error {
		var item about.Skill
		err := tx.Where("key = ?", seed.Key).First(&item).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			item = about.Skill{
				Key:       seed.Key,
				Category:  seed.Category,
				Level:     seed.Level,
				SortOrder: seed.SortOrder,
				Active:    seed.Active,
			}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}

		updates := map[string]any{
			"key":        seed.Key,
			"category":   seed.Category,
			"level":      seed.Level,
			"sort_order": seed.SortOrder,
			"active":     seed.Active,
		}
		if err := tx.Model(&item).Updates(updates).Error; err != nil {
			return err
		}
		return replaceSkillTranslations(tx, item.ID, seed.Translations)
	})
}

func upsertInterest(database *gorm.DB, seed interestSeed) error {
	return database.Transaction(func(tx *gorm.DB) error {
		var item about.Interest
		err := tx.Where("key = ?", seed.Key).First(&item).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			item = about.Interest{Key: seed.Key}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}

		updates := map[string]any{
			"key":        seed.Key,
			"sort_order": seed.SortOrder,
			"active":     seed.Active,
		}
		if err := tx.Model(&item).Updates(updates).Error; err != nil {
			return err
		}
		return replaceInterestTranslations(tx, item.ID, seed.Translations)
	})
}

func replaceExperienceTranslations(tx *gorm.DB, experienceID uint64, translations []about.ExperienceTranslation) error {
	if err := tx.Where("experience_id = ?", experienceID).Delete(&about.ExperienceTranslation{}).Error; err != nil {
		return err
	}
	for i := range translations {
		translations[i].ExperienceID = experienceID
	}
	if len(translations) == 0 {
		return nil
	}
	return tx.Create(&translations).Error
}

func replaceEducationTranslations(tx *gorm.DB, educationID uint64, translations []about.EducationTranslation) error {
	if err := tx.Where("education_id = ?", educationID).Delete(&about.EducationTranslation{}).Error; err != nil {
		return err
	}
	for i := range translations {
		translations[i].EducationID = educationID
	}
	if len(translations) == 0 {
		return nil
	}
	return tx.Create(&translations).Error
}

func replaceSkillTranslations(tx *gorm.DB, skillID uint64, translations []about.SkillTranslation) error {
	if err := tx.Where("skill_id = ?", skillID).Delete(&about.SkillTranslation{}).Error; err != nil {
		return err
	}
	for i := range translations {
		translations[i].SkillID = skillID
	}
	if len(translations) == 0 {
		return nil
	}
	return tx.Create(&translations).Error
}

func replaceInterestTranslations(tx *gorm.DB, interestID uint64, translations []about.InterestTranslation) error {
	if err := tx.Where("interest_id = ?", interestID).Delete(&about.InterestTranslation{}).Error; err != nil {
		return err
	}
	for i := range translations {
		translations[i].InterestID = interestID
	}
	if len(translations) == 0 {
		return nil
	}
	return tx.Create(&translations).Error
}

func parseDate(value string) (time.Time, error) {
	return time.Parse("2006-01-02", value)
}

func parseOptionalDate(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := parseDate(value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
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

func experienceSeeds() []experienceSeed {
	return []experienceSeed{
		{
			Company:      "Sielaff Software und Service GmbH & Co. KG",
			StartDate:    "2025-08-01",
			EndDate:      "2026-01-31",
			IsCurrent:    false,
			SortOrder:    10,
			Active:       true,
			Technologies: []string{"Django", "DRF", "React", "Next.js", "Docker", "GitLab CI/CD", "WebSockets"},
			Translations: []about.ExperienceTranslation{
				{
					Locale:  "de",
					Role:    "Fullstack-Webentwickler",
					Summary: "Weiterentwicklung interner Webanwendungen im Industrie-4.0-Umfeld.",
					Highlights: []string{
						"Backend-Arbeit mit Django und DRF, Frontend mit React und Next.js.",
						"Deployment und Containerisierung mit Docker und GitLab CI/CD.",
						"Echtzeit-Kommunikation ueber WebSockets und Datenbankoptimierung.",
					},
				},
				{
					Locale:  "en",
					Role:    "Fullstack Web Developer",
					Summary: "Further development of internal web applications in the Industry 4.0 environment.",
					Highlights: []string{
						"Backend work with Django and DRF, frontend work with React and Next.js.",
						"Deployment and containerization with Docker and GitLab CI/CD.",
						"Realtime communication via WebSockets and database optimization.",
					},
				},
			},
		},
		{
			Company:   "Ausbildung bei Sielaff",
			StartDate: "2022-09-01",
			EndDate:   "2025-07-31",
			IsCurrent: false,
			SortOrder: 20,
			Active:    true,
			Translations: []about.ExperienceTranslation{
				{
					Locale:  "de",
					Role:    "Fachinformatiker fuer Anwendungsentwicklung",
					Summary: "Ausbildung mit Schwerpunkt Anwendungsentwicklung und Webentwicklung.",
				},
				{
					Locale:  "en",
					Role:    "IT Specialist for Application Development",
					Summary: "Apprenticeship focused on application development and web development.",
				},
			},
		},
	}
}

func educationSeeds() []educationSeed {
	return []educationSeed{
		{
			Institution: "Andreas Gordon Schule",
			Location:    "Erfurt",
			StartDate:   "2022-09-01",
			EndDate:     "2025-07-31",
			IsCurrent:   false,
			SortOrder:   10,
			Active:      true,
			Translations: []about.EducationTranslation{
				{
					Locale:  "de",
					Title:   "Berufsschule",
					Summary: "Fachinformatiker fuer Anwendungsentwicklung.",
					Highlights: []string{
						"Bester Absolvent der IHK Suedthueringen 2025.",
					},
				},
				{
					Locale:  "en",
					Title:   "Vocational School",
					Summary: "IT Specialist for Application Development.",
					Highlights: []string{
						"Best graduate of IHK South Thuringia 2025.",
					},
				},
			},
		},
		{
			Institution: "Goetheschule Ilmenau",
			Location:    "Ilmenau",
			StartDate:   "2013-08-01",
			EndDate:     "2021-07-31",
			IsCurrent:   false,
			SortOrder:   20,
			Active:      true,
			Translations: []about.EducationTranslation{
				{Locale: "de", Title: "Schulabschluss", Summary: "Allgemeinbildende Schule bis 2021."},
				{Locale: "en", Title: "School Graduation", Summary: "General education until 2021."},
			},
		},
	}
}

func skillSeeds() []skillSeed {
	return []skillSeed{
		skillSeedEntry("python", about.SkillCategoryBackend, about.SkillLevelProfessional, 10, "Python", "Backend-Logik, Automatisierung und kleine interne Tools."),
		skillSeedEntry("django", about.SkillCategoryBackend, about.SkillLevelProfessional, 20, "Django", "Produktive Webanwendungen und interne Business-Systeme."),
		skillSeedEntry("react", about.SkillCategoryFrontend, about.SkillLevelProfessional, 30, "React", "Interaktive Oberflaechen, State-Flows und komponentenbasierte UIs."),
		skillSeedEntry("nextjs", about.SkillCategoryFrontend, about.SkillLevelProfessional, 40, "Next.js", "App Router, Server Components und datengetriebene Seiten."),
		skillSeedEntry("typescript", about.SkillCategoryFrontend, about.SkillLevelProfessional, 50, "TypeScript", "Typisierte Komponenten, API-Modelle und sichere Refactors."),
		skillSeedEntry("docker", about.SkillCategoryDevOps, about.SkillLevelComfortable, 60, "Docker", "Containerisierung und lokale Entwicklungsumgebungen."),
		skillSeedEntry("postgresql", about.SkillCategoryBackend, about.SkillLevelComfortable, 70, "PostgreSQL", "Relationale Datenmodelle, Abfragen und Optimierungen."),
	}
}

func skillSeedEntry(key string, category about.SkillCategory, level about.SkillLevel, sortOrder int32, name string, deSummary string) skillSeed {
	return skillSeed{
		Key:       key,
		Category:  category,
		Level:     level,
		SortOrder: sortOrder,
		Active:    true,
		Translations: []about.SkillTranslation{
			{Locale: "de", Name: name, Summary: deSummary},
			{Locale: "en", Name: name, Summary: deSummary},
		},
	}
}

func interestSeeds() []interestSeed {
	return []interestSeed{
		{
			Key:       "systems",
			SortOrder: 10,
			Active:    true,
			Translations: []about.InterestTranslation{
				{Locale: "de", Name: "Systeme", Summary: "Betriebssysteme, Netzwerke, Infrastruktur und wie Software im Ganzen zusammenhaengt."},
				{Locale: "en", Name: "Systems", Summary: "Operating systems, networks, infrastructure, and how software fits together."},
			},
		},
		{
			Key:       "design",
			SortOrder: 20,
			Active:    true,
			Translations: []about.InterestTranslation{
				{Locale: "de", Name: "Interface Design", Summary: "Ruhige, nutzbare Oberflaechen mit klarer Struktur und guter Interaktion."},
				{Locale: "en", Name: "Interface Design", Summary: "Quiet, usable interfaces with clear structure and good interaction."},
			},
		},
		{
			Key:       "writing",
			SortOrder: 30,
			Active:    true,
			Translations: []about.InterestTranslation{
				{Locale: "de", Name: "Schreiben", Summary: "Notizen, technische Dokumentation und Texte, die Gedanken praezise machen."},
				{Locale: "en", Name: "Writing", Summary: "Notes, technical documentation, and writing that makes thoughts precise."},
			},
		},
	}
}

func printUsage() {
	fmt.Println("usage:")
	fmt.Println("  seed about   seed about page data")
	fmt.Println("  seed all     seed all core fixtures")
}
