-- SQLite version
-- set your range: from_date inclusive, to_date exclusive (months before to_date)

WITH RECURSIVE months(m) AS (
    SELECT date('2022-12-01')
    UNION ALL
    SELECT date(m, '+1 month')
    FROM months
    WHERE m < date(date('2026-12-31'), '-1 month')
),
cat_month AS (
    SELECT
        t.category AS category,
        strftime('%Y-%m', t.date) AS ym,
        SUM(t.amount) AS amt
    FROM transactions t
    WHERE t.date >= date('2022-12-01') AND t.date < date('2026-12-31')
    GROUP BY t.category, ym
)
SELECT
    c.id,
    c.code,
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
          ON cm.category = c.code
         AND cm.ym = mon.ym
    ) AS monthly
FROM categories c
ORDER BY c.id;
