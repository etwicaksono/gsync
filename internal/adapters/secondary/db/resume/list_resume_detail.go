package resume

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/model"
	"github.com/etwicaksono/data-transaksi-service/internal/constants"
	"gorm.io/gorm"
)

func (a adapter) ListResumeDetail(ctx context.Context, tx *gorm.DB, filter entity.VResumeDetailListFilter) (result []entity.Resume, err error) {
	var resumeSQL = fmt.Sprintf(`
			WITH RECURSIVE months(m) AS (
			  SELECT DATE(?)                           -- @from (inclusive)
			  UNION ALL
			  SELECT DATE_ADD(m, INTERVAL 1 MONTH)
			  FROM months
			  WHERE m < DATE_SUB(DATE(?), INTERVAL 1 MONTH)   -- @to (exclusive)
			),
			cat_month AS (
			  SELECT
				t.category AS category,
				DATE_FORMAT(t.date, '%%Y-%%m') AS ym,
				SUM(t.amount) AS amt
			  FROM %s t
			  WHERE t.date >= DATE(?) AND t.date < DATE(?)    -- same window
			  GROUP BY t.category, ym
			)
			SELECT
				c.id,
				c.code,
				CAST(CONCAT(
						'[',
						GROUP_CONCAT(
							JSON_OBJECT(
								'ym',  DATE_FORMAT(mon.m, '%%Y-%%m'),
								'amt', COALESCE(cm.amt, 0)
							) ORDER BY mon.m SEPARATOR ','
						),
						']'
					 ) AS JSON) AS monthly
			FROM %s c
			CROSS JOIN months mon
			LEFT JOIN cat_month cm
			  ON cm.category = c.code
			 AND cm.ym = DATE_FORMAT(mon.m, '%%Y-%%m')
			GROUP BY c.id, c.code
			ORDER BY c.id;
			`, model.Transaction{}.TableName(), model.Category{}.TableName())

	var rows []model.Resume
	// placeholders order: DATE(?), DATE(?), DATE(?), DATE(?)
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
		result[i] = item.ToEntity()
	}

	return
}
