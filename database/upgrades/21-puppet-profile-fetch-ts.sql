-- v21 (compatible with v18+): Add profile fetch timestamp for puppets
ALTER TABLE puppet ADD profile_fetched_at BIGINT;
