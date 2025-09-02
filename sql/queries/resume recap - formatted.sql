-- set your range: @from inclusive, @to exclusive (months before @to)
SET @from := DATE('2022-12-01');
SET @to   := DATE('2026-12-31');
-- SET @from := DATE('2025-08-01');
-- SET @to   := DATE('2025-08-31');

-- avoid truncation in GROUP_CONCAT
SET SESSION group_concat_max_len = 1000000;

WITH RECURSIVE months(m) AS (
    SELECT @from
    UNION ALL
    SELECT DATE_ADD(m, INTERVAL 1 MONTH)
    FROM months
    WHERE m < DATE_SUB(@to, INTERVAL 1 MONTH)
)
SELECT GROUP_CONCAT(
               CONCAT(
                       'rp_fmt(SUM(CASE WHEN t.date >= ''', DATE_FORMAT(m, '%Y-%m-01'),
                       ''' AND t.date < ''', DATE_FORMAT(DATE_ADD(m, INTERVAL 1 MONTH), '%Y-%m-01'),
                       ''' THEN t.amount END)) AS `', DATE_FORMAT(m, '%Y %M'), '`'
               )
                   ORDER BY m
  SEPARATOR ', '
       ) INTO @cols
FROM months;

SET @sql = CONCAT(
  'SELECT COALESCE(p.id, c.id) AS group_id,
          COALESCE(p.code, c.code) AS `group`, ',
          @cols, '
   FROM categories AS c
   LEFT JOIN categories AS p
     ON p.name = c.parent
   LEFT JOIN transactions AS t
     ON t.category = c.code
   WHERE t.date IS NOT NULL
   GROUP BY group_id, `group`
   ORDER BY group_id;'
);

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
