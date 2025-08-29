SELECT
    a.code AS account,
    rp_fmt(SUM(m.amt)) AS current_amount
FROM accounts a
-- keep only accounts that appear in transactions (like your original base "tx")
         JOIN (SELECT account FROM transactions GROUP BY account) t
              ON t.account = a.code
         LEFT JOIN (
    -- + transactions
    SELECT account,  SUM(amount)        AS amt FROM transactions GROUP BY account
    UNION ALL
    -- - credits
    SELECT account, -SUM(amount)              FROM credits      GROUP BY account
    UNION ALL
    -- + debts
    SELECT account,  SUM(amount)              FROM debts        GROUP BY account
    UNION ALL
    -- + transfers in
    SELECT to_account   AS account,  SUM(amount) FROM transfers GROUP BY to_account
    UNION ALL
    -- - transfers out
    SELECT from_account AS account, -SUM(amount) FROM transfers GROUP BY from_account
) AS m
                   ON m.account = a.code
GROUP BY a.id, a.code
ORDER BY a.id ASC;