ALTER TABLE about_skills DROP CONSTRAINT IF EXISTS chk_about_skills_category;
ALTER TABLE about_skills ADD CONSTRAINT chk_about_skills_category
    CHECK (category IN ('backend', 'frontend', 'devops', 'tooling', 'infrastructure', 'other'));
