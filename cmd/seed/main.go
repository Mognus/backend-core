package main

import (
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

func seedExperiences(database *gorm.DB, seeds []experienceSeed) {
	for _, seed := range seeds {
		if err := upsertExperience(database, seed); err != nil {
			log.Fatalf("seed experience %s: %v", seed.Company, err)
		}
		fmt.Printf("seeded experience %s\n", seed.Company)
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
			"technologies": seed.Technologies,
		}
		if err := tx.Model(&item).Updates(updates).Error; err != nil {
			return err
		}
		return replaceExperienceTranslations(tx, item.ID, seed.Translations)
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
