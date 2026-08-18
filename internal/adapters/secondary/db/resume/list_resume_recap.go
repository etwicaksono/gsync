package resume

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/model"
	"github.com/etwicaksono/data-transaksi-service/internal/constants"
	"gorm.io/gorm"
)

func (a adapter) ListResumeRecap(ctx context.Context, tx *gorm.DB, filter entity.VResumeDetailListFilter) (result []entity.Resume, err error) {
	transactionTable := model.Transaction{}.TableName()
	categoryTable := model.Category{}.TableName()

	resumeSQL := strings.TrimSpace(fmt.Sprintf(`
		WITH RECURSIVE months(m) AS (
		  SELECT date(?)
		  UNION ALL
		  SELECT date(m, '+1 month')
		  FROM months
		  WHERE m < date(date(?), '-1 month')
		),

		cat_groups AS (
		  SELECT DISTINCT
			COALESCE(p.id,   c.id)   AS id,
			COALESCE(p.code, c.code) AS code
		  FROM %s c
		  LEFT JOIN %s p ON p.name = c.parent
		),

		cat_month AS (
		  SELECT
			COALESCE(p.code, c.code)     AS gcode,
			strftime('%%Y-%%m', t.date)    AS ym,
			SUM(t.amount)                AS amt
		  FROM %s t
		  JOIN %s c ON t.category = c.code
		  LEFT JOIN %s p ON p.name = c.parent
		  WHERE t.date >= date(?) AND t.date < date(?)
		  GROUP BY gcode, ym
		)

		SELECT
		  g.id,
		  g.code AS code,
		  (
			SELECT json_group_array(
				json_object(
					'ym',  mon.ym,
					'amt', COALESCE(cm.amt, 0)
				)
			)
			FROM (SELECT strftime('%%Y-%%m', m) AS ym FROM months) mon
			LEFT JOIN cat_month cm
			  ON cm.gcode = g.code
			 AND cm.ym = mon.ym
		  ) AS monthly
		FROM cat_groups g
		ORDER BY g.id;
		`,
		categoryTable,
		categoryTable,
		transactionTable,
		categoryTable,
		categoryTable,
	))

	var rows []model.Resume
	if err = tx.WithContext(ctx).
		Raw(resumeSQL,
			filter.StartDate.Format(time.DateOnly),
			filter.EndDate.Format(time.DateOnly),
			filter.StartDate.Format(time.DateOnly),
			filter.EndDate.Format(time.DateOnly),
		).Scan(&rows).Error; err != nil {
		slog.ErrorContext(ctx, "failed to list resume recap", slog.String(constants.Error, err.Error()))
		return nil, err
	}

	result = make([]entity.Resume, len(rows))
	for i, item := range rows {
		item.Monthly.SortByYM()
		result[i] = item.ToEntity()
	}

	return
}
