-- PostgreSQL version
-- set your range: from_date inclusive, to_date exclusive (months before to_date)

WITH RECURSIVE months(m) AS (
    SELECT '2022-12-01'::DATE
    UNION ALL
    SELECT m + INTERVAL '1 month'
    FROM months
    WHERE m < '2026-12-31'::DATE - INTERVAL '1 month'
),

-- distinct parent groups (parent if exists, else the category itself)
cat_groups AS (
    SELECT DISTINCT
        COALESCE(p.id, c.id) AS id,
        COALESCE(p.code, c.code) AS code
    FROM categories c
    LEFT JOIN categories p ON p.name = c.parent
),

-- sums by group code and year-month within the window
cat_month AS (
    SELECT
        COALESCE(p.code, c.code) AS gcode,
        TO_CHAR(t.date, 'YYYY-MM') AS ym,
        SUM(t.amount) AS amt
    FROM transactions t
    JOIN categories c ON t.category = c.code
    LEFT JOIN categories p ON p.name = c.parent
    WHERE t.date >= '2022-12-01'::DATE AND t.date < '2026-12-31'::DATE
    GROUP BY gcode, ym
)

SELECT
    g.id,
    g.code AS code,
    (
        SELECT JSONB_AGG(
            JSONB_BUILD_OBJECT(
                'ym', TO_CHAR(mon2.m, 'YYYY-MM'),
                'amt', COALESCE(cm2.amt, 0)
            ) ORDER BY mon2.m
        )
        FROM months mon2
        LEFT JOIN cat_month cm2
          ON cm2.gcode = g.code
         AND cm2.ym = TO_CHAR(mon2.m, 'YYYY-MM')
    ) AS monthly
FROM cat_groups g
GROUP BY g.id, g.code
ORDER BY g.id;
