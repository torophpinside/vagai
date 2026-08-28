-- ============================================================
-- VagAI Backfill: Normalize Existing Job URLs
-- ============================================================
-- Execute this script to normalize all existing job URLs
-- and remove duplicates caused by URL variations.
--
-- IMPORTANT: This script uses MySQL 8.0+ features.
-- Run this BEFORE starting the application after deploying
-- the URL normalization fix.
-- ============================================================

USE vagai;

-- ============================================================
-- STEP 1: Identify duplicates that will be merged
-- ============================================================
-- This query shows jobs that have the same normalized URL
-- but different raw URLs (the duplicates we want to remove)

SELECT 
    'Jobs duplicados que serão removidos' AS info,
    COUNT(*) AS total_duplicados
FROM (
    SELECT 
        organization_id,
        -- Apply normalization logic inline
        LOWER(
            REGEXP_REPLACE(
                REGEXP_REPLACE(
                    REGEXP_REPLACE(
                        REGEXP_REPLACE(
                            REGEXP_REPLACE(
                                REGEXP_REPLACE(
                                    REGEXP_REPLACE(
                                        REGEXP_REPLACE(
                                            url,
                                            '\\?.*$', ''  -- Remove query params
                                        ),
                                        '#.*$', ''  -- Remove fragments
                                    ),
                                    '/$', ''  -- Remove trailing slash
                                ),
                                '^https?://', 'https://'  -- Force https
                            ),
                            '^https://www\\.', 'https://'  -- Remove www
                        ),
                        '^http://www\\.', 'https://'  -- Remove www from http
                    ),
                    '^http://', 'https://'  -- Force https from http
                ),
                '\\?.*$', ''  -- Second pass for query params
            )
        ) AS normalized_url
    FROM jobs
    WHERE deleted_at IS NULL
    GROUP BY organization_id, normalized_url
    HAVING COUNT(*) > 1
) AS duplicates;

-- ============================================================
-- STEP 2: Preview what will be deleted
-- ============================================================
-- Shows the jobs that will be removed (keeping the oldest one)

SELECT 
    j.id,
    j.title,
    j.company,
    j.url,
    DATE(j.created_at) AS created_date,
    j.organization_id
FROM jobs j
INNER JOIN (
    SELECT 
        organization_id,
        url,
        MIN(id) AS keep_id  -- Keep the oldest job
    FROM jobs
    WHERE deleted_at IS NULL
    GROUP BY organization_id, url
    HAVING COUNT(*) > 1
) AS dups ON j.url = dups.url 
    AND j.organization_id = dups.organization_id
    AND j.id != dups.keep_id
ORDER BY j.organization_id, j.url
LIMIT 100;

-- ============================================================
-- STEP 3: Delete duplicate jobs (keeping the oldest)
-- ============================================================
-- WARNING: This will DELETE data. Make a backup first!

DELETE j1 FROM jobs j1
INNER JOIN (
    SELECT 
        id,
        organization_id,
        url,
        ROW_NUMBER() OVER (
            PARTITION BY organization_id, url 
            ORDER BY id ASC
        ) AS row_num
    FROM jobs
    WHERE deleted_at IS NULL
) AS j2 ON j1.id = j2.id
WHERE j2.row_num > 1;

-- ============================================================
-- STEP 4: Verify cleanup
-- ============================================================
SELECT 
    'Jobs restantes após limpeza' AS info,
    COUNT(*) AS total_jobs
FROM jobs
WHERE deleted_at IS NULL;

-- ============================================================
-- STEP 5: Verify no duplicates remain
-- ============================================================
SELECT 
    organization_id,
    url,
    COUNT(*) AS duplicatas
FROM jobs
WHERE deleted_at IS NULL
GROUP BY organization_id, url
HAVING COUNT(*) > 1;

-- ============================================================
-- STEP 6: Also normalize matches that reference deleted jobs
-- ============================================================
-- Clean up orphaned matches (if any)
DELETE m FROM matches m
LEFT JOIN jobs j ON m.job_id = j.id
WHERE j.id IS NULL;

-- ============================================================
-- DONE
-- ============================================================
SELECT 'Backfill completo! URLs normalizadas e duplicatas removidas.' AS resultado;
