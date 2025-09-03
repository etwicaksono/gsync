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

func (a adapter) ListResumeRecap(ctx context.Context, tx *gorm.DB, filter entity.VResumeDetailListFilter) (result []entity.Resume, err error) {
	var resumeSQL = fmt.Sprintf(`
			WITH RECURSIVE months(m) AS (
			  SELECT DATE(?)
			  UNION ALL
			  SELECT DATE_ADD(m, INTERVAL 1 MONTH)
			  FROM months
			  WHERE m < DATE_SUB(DATE(?), INTERVAL 1 MONTH)
			),
			
			-- distinct parent groups (parent if exists, else the category itself)
			cat_groups AS (
			  SELECT DISTINCT
				COALESCE(p.id,   c.id)   AS id,
				COALESCE(p.code, c.code) AS code
			  FROM %s c
			  LEFT JOIN %s p ON p.name = c.parent
			),
			
			-- sums by group code and year-month within the window
			cat_month AS (
			  SELECT
				COALESCE(p.code, c.code)       AS gcode,
				DATE_FORMAT(t.date, '%%Y-%%m')   AS ym,
				SUM(t.amount)                  AS amt
			  FROM %s t
			  JOIN %s c ON t.category = c.code
			  LEFT JOIN %s p ON p.name = c.parent
			  WHERE t.date >= DATE(?) AND t.date < DATE(?)
			  GROUP BY gcode, ym
			)
			
			SELECT
			  g.id,
			  g.code AS code,
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
			FROM cat_groups g
			CROSS JOIN months mon
			LEFT JOIN cat_month cm
			  ON cm.gcode = g.code
			 AND cm.ym    = DATE_FORMAT(mon.m, '%%Y-%%m')
			GROUP BY g.id, g.code
			ORDER BY g.id;
			`,
		model.Category{}.TableName(),
		model.Category{}.TableName(),
		model.Transaction{}.TableName(),
		model.Category{}.TableName(),
		model.Category{}.TableName(),
	)

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
