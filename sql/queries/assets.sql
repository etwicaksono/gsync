-- SQLite version
-- rp_fmt() was a PostgreSQL plpgsql helper; SQLite has no user-defined
-- functions from plain SQL, so the raw sum is returned instead and formatting
-- is left to the caller.
SELECT
    a.code AS account,
    SUM(m.amt) AS current_amount
FROM accounts a
-- keep only accounts that appear in transactions (like your original base "tx")
         JOIN (SELECT account FROM transactions GROUP BY account) t
              ON t.account = a.code
         LEFT JOIN (
    -- + transactions
    SELECT account AS account,  SUM(amount)  AS amt FROM transactions GROUP BY account
    UNION ALL
    -- - credits
    SELECT account AS account, -SUM(amount)  AS amt FROM credits      GROUP BY account
    UNION ALL
    -- + debts
    SELECT account AS account,  SUM(amount)  AS amt FROM debts        GROUP BY account
    UNION ALL
    -- + transfers in
    SELECT to_account   AS account,  SUM(amount) AS amt FROM transfers GROUP BY to_account
    UNION ALL
    -- - transfers out
    SELECT from_account AS account, -SUM(amount) AS amt FROM transfers GROUP BY from_account
) AS m
                   ON m.account = a.code
GROUP BY a.id, a.code
ORDER BY a.id ASC;
