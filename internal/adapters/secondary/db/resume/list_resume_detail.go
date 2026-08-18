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

func (a adapter) ListResumeDetail(ctx context.Context, tx *gorm.DB, filter entity.VResumeDetailListFilter) (result []entity.Resume, err error) {
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
		cat_month AS (
		  SELECT
			t.category AS category,
			strftime('%%Y-%%m', t.date) AS ym,
			SUM(t.amount) AS amt
		  FROM %s t
		  WHERE t.date >= date(?) AND t.date < date(?)
		  GROUP BY t.category, ym
		)
		SELECT
			c.id,
			c.code,
			(
				SELECT json_group_array(
					json_object(
						'ym',  mon.ym,
						'amt', COALESCE(cm.amt, 0)
					)
				)
				FROM (SELECT strftime('%%Y-%%m', m) AS ym FROM months) mon
				LEFT JOIN cat_month cm
				  ON cm.category = c.code
				 AND cm.ym = mon.ym
			) AS monthly
		FROM %s c
		ORDER BY c.id;
		`, transactionTable, categoryTable))

	var rows []model.Resume
	if err = tx.WithContext(ctx).
		Raw(resumeSQL,
			filter.StartDate.Format(time.DateOnly),
			filter.EndDate.Format(time.DateOnly),
			filter.StartDate.Format(time.DateOnly),
			filter.EndDate.Format(time.DateOnly),
		).Scan(&rows).Error; err != nil {
		slog.ErrorContext(ctx, "failed to list resume detail", slog.String(constants.Error, err.Error()))
		return nil, err
	}

	result = make([]entity.Resume, len(rows))
	for i, item := range rows {
		item.Monthly.SortByYM()
		result[i] = item.ToEntity()
	}

	return
}
