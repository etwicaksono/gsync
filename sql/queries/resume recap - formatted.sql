-- SQLite version
-- set your range: from_date inclusive, to_date exclusive (months before to_date)

WITH RECURSIVE months(m) AS (
    SELECT date('2022-12-01')
    UNION ALL
    SELECT date(m, '+1 month')
    FROM months
    WHERE m < date(date('2026-12-31'), '-1 month')
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
        strftime('%Y-%m', t.date) AS ym,
        SUM(t.amount) AS amt
    FROM transactions t
    JOIN categories c ON t.category = c.code
    LEFT JOIN categories p ON p.name = c.parent
    WHERE t.date >= date('2022-12-01') AND t.date < date('2026-12-31')
    GROUP BY gcode, ym
)

SELECT
    g.id,
    g.code AS code,
    (
        -- SQLite does not support ORDER BY inside an aggregate, so the month
        -- series is ordered in the subquery that feeds json_group_array.
        SELECT json_group_array(
            json_object(
                'ym',  mon.ym,
                'amt', COALESCE(cm.amt, 0)
            )
        )
        FROM (SELECT strftime('%Y-%m', m) AS ym FROM months ORDER BY m) mon
        LEFT JOIN cat_month cm
          ON cm.gcode = g.code
         AND cm.ym = mon.ym
    ) AS monthly
FROM cat_groups g
ORDER BY g.id;
