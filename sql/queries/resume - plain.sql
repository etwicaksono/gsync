-- set your range: @from inclusive, @to exclusive (months before @to)
SET @from := DATE('2022-12-01');
SET @to   := DATE('2025-09-01');   -- up to 2025-08

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
                       'SUM(CASE WHEN t.date >= ''', DATE_FORMAT(m, '%Y-%m-01'),
                       ''' AND t.date < ''', DATE_FORMAT(DATE_ADD(m, INTERVAL 1 MONTH), '%Y-%m-01'),
                       ''' THEN t.amount END) AS `', DATE_FORMAT(m, '%Y %M'), '`'
               )
                   ORDER BY m
  SEPARATOR ', '
       ) INTO @cols
FROM months;

SET @sql = CONCAT(
  'SELECT c.id, c.code, ', @cols, '
   FROM categories c
   LEFT JOIN transactions t ON t.category = c.code
   GROUP BY c.id, c.code
   ORDER BY c.id;'
);

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
