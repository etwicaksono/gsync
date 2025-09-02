with recursive `months` (`m`) as (
    select
        cast('2022-12-01' as date) AS `DATE('2022-12-01')`
    union all
    select
        (`months`.`m` + interval 1 month) AS `DATE_ADD(m, INTERVAL 1 MONTH)`
    from
        `months`
    where
        (
            `months`.`m` < (
                cast('2026-12-31' as date) - interval 1 month
                )
            )
),
               `cat_month` as (
                   select
                       `t`.`category` AS `category`,
                       date_format(`t`.`date`, '%Y-%m') AS `ym`,
                       sum(`t`.`amount`) AS `amt`
                   from
                       `transactions` `t`
                   where
                       (
                           (
                               `t`.`date` >= cast('2022-12-01' as date)
                               )
                               and (
                               `t`.`date` < cast('2026-12-31' as date)
                               )
                           )
                   group by
                       `t`.`category`,
                       `ym`
               )
select
    `c`.`id` AS `id`,
    `c`.`code` AS `code`,
    json_objectagg(
            date_format(`mon`.`m`, '%Y-%m'),
            coalesce(`cm`.`amt`, 0)
    ) AS `monthly`
from
    (
        (
            `categories` `c`
                join `months` `mon`
            )
            left join `cat_month` `cm` on(
            (
                (`cm`.`category` = `c`.`code`)
                    and (
                    `cm`.`ym` = date_format(`mon`.`m`, '%Y-%m')
                    )
                )
            )
        )
group by
    `c`.`id`,
    `c`.`code`
order by
    `c`.`id`
