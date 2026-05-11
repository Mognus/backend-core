CREATE TABLE about_experiences (
    id           BIGSERIAL PRIMARY KEY,
    company      VARCHAR(160) NOT NULL,
    location     VARCHAR(160),
    start_date   DATE         NOT NULL,
    end_date     DATE,
    is_current   BOOLEAN      NOT NULL DEFAULT FALSE,
    sort_order   INTEGER      NOT NULL DEFAULT 0,
    active       BOOLEAN      NOT NULL DEFAULT TRUE,
    technologies JSONB        NOT NULL DEFAULT '[]'::jsonb,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_about_experiences_dates
        CHECK (end_date IS NULL OR end_date >= start_date),
    CONSTRAINT chk_about_experiences_current
        CHECK ((is_current = TRUE AND end_date IS NULL) OR is_current = FALSE),
    CONSTRAINT chk_about_experiences_technologies_array
        CHECK (jsonb_typeof(technologies) = 'array')
);

CREATE TABLE about_experience_translations (
    id            BIGSERIAL PRIMARY KEY,
    experience_id BIGINT       NOT NULL,
    locale        VARCHAR(8)   NOT NULL,
    role          VARCHAR(180) NOT NULL,
    summary       TEXT,
    highlights    JSONB        NOT NULL DEFAULT '[]'::jsonb,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_about_experience_translations_experience FOREIGN KEY (experience_id)
        REFERENCES about_experiences (id)
        ON DELETE CASCADE,
    CONSTRAINT about_experience_translations_experience_locale_key UNIQUE (experience_id, locale),
    CONSTRAINT chk_about_experience_translations_highlights_array
        CHECK (jsonb_typeof(highlights) = 'array')
);

CREATE TABLE about_education (
    id          BIGSERIAL PRIMARY KEY,
    institution VARCHAR(180) NOT NULL,
    location    VARCHAR(160),
    start_date  DATE         NOT NULL,
    end_date    DATE,
    is_current  BOOLEAN      NOT NULL DEFAULT FALSE,
    sort_order  INTEGER      NOT NULL DEFAULT 0,
    active      BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_about_education_dates
        CHECK (end_date IS NULL OR end_date >= start_date),
    CONSTRAINT chk_about_education_current
        CHECK ((is_current = TRUE AND end_date IS NULL) OR is_current = FALSE)
);

CREATE TABLE about_education_translations (
    id           BIGSERIAL PRIMARY KEY,
    education_id BIGINT       NOT NULL,
    locale       VARCHAR(8)   NOT NULL,
    title        VARCHAR(180) NOT NULL,
    summary      TEXT,
    highlights   JSONB        NOT NULL DEFAULT '[]'::jsonb,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_about_education_translations_education FOREIGN KEY (education_id)
        REFERENCES about_education (id)
        ON DELETE CASCADE,
    CONSTRAINT about_education_translations_education_locale_key UNIQUE (education_id, locale),
    CONSTRAINT chk_about_education_translations_highlights_array
        CHECK (jsonb_typeof(highlights) = 'array')
);

CREATE TABLE about_skills (
    id         BIGSERIAL PRIMARY KEY,
    key        VARCHAR(120) NOT NULL,
    category   VARCHAR(40)  NOT NULL,
    level      VARCHAR(40)  NOT NULL,
    sort_order INTEGER      NOT NULL DEFAULT 0,
    active     BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT about_skills_key_key UNIQUE (key),
    CONSTRAINT chk_about_skills_category
        CHECK (category IN ('backend', 'frontend', 'devops', 'tooling', 'other')),
    CONSTRAINT chk_about_skills_level
        CHECK (level IN ('learning', 'comfortable', 'strong', 'professional'))
);

CREATE TABLE about_skill_translations (
    id         BIGSERIAL PRIMARY KEY,
    skill_id   BIGINT       NOT NULL,
    locale     VARCHAR(8)   NOT NULL,
    name       VARCHAR(120) NOT NULL,
    summary    TEXT,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_about_skill_translations_skill FOREIGN KEY (skill_id)
        REFERENCES about_skills (id)
        ON DELETE CASCADE,
    CONSTRAINT about_skill_translations_skill_locale_key UNIQUE (skill_id, locale)
);

CREATE TABLE about_interests (
    id         BIGSERIAL PRIMARY KEY,
    key        VARCHAR(120) NOT NULL,
    sort_order INTEGER      NOT NULL DEFAULT 0,
    active     BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT about_interests_key_key UNIQUE (key)
);

CREATE TABLE about_interest_translations (
    id          BIGSERIAL PRIMARY KEY,
    interest_id BIGINT       NOT NULL,
    locale      VARCHAR(8)   NOT NULL,
    name        VARCHAR(120) NOT NULL,
    summary     TEXT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_about_interest_translations_interest FOREIGN KEY (interest_id)
        REFERENCES about_interests (id)
        ON DELETE CASCADE,
    CONSTRAINT about_interest_translations_interest_locale_key UNIQUE (interest_id, locale)
);

CREATE INDEX idx_about_experiences_active_sort
    ON about_experiences (active, sort_order, start_date DESC, id);
CREATE INDEX idx_about_experience_translations_locale
    ON about_experience_translations (locale);

CREATE INDEX idx_about_education_active_sort
    ON about_education (active, sort_order, start_date DESC, id);
CREATE INDEX idx_about_education_translations_locale
    ON about_education_translations (locale);

CREATE INDEX idx_about_skills_active_sort
    ON about_skills (active, sort_order, id);
CREATE INDEX idx_about_skills_category
    ON about_skills (category);
CREATE INDEX idx_about_skill_translations_locale
    ON about_skill_translations (locale);

CREATE INDEX idx_about_interests_active_sort
    ON about_interests (active, sort_order, id);
CREATE INDEX idx_about_interest_translations_locale
    ON about_interest_translations (locale);
