-- PostgreSQL version
-- set your range: from_date inclusive, to_date exclusive (months before to_date)

WITH RECURSIVE months(m) AS (
    SELECT '2022-12-01'::DATE
    UNION ALL
    SELECT m + INTERVAL '1 month'
    FROM months
    WHERE m < '2026-12-31'::DATE - INTERVAL '1 month'
),
cat_month AS (
    SELECT
        t.category AS category,
        TO_CHAR(t.date, 'YYYY-MM') AS ym,
        SUM(t.amount) AS amt
    FROM transactions t
    WHERE t.date >= '2022-12-01'::DATE AND t.date < '2026-12-31'::DATE
    GROUP BY t.category, ym
)
SELECT
    c.id,
    c.code,
    (
        SELECT JSONB_AGG(
            JSONB_BUILD_OBJECT(
                'ym',  TO_CHAR(mon2.m, 'YYYY-MM'),
                'amt', COALESCE(cm2.amt, 0)
            ) ORDER BY mon2.m
        )
        FROM months mon2
        LEFT JOIN cat_month cm2
          ON cm2.category = c.code
         AND cm2.ym = TO_CHAR(mon2.m, 'YYYY-MM')
    ) AS monthly
FROM categories c
GROUP BY c.id, c.code
ORDER BY c.id;
