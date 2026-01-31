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
			  SELECT ?::DATE
			  UNION ALL
			  SELECT m + INTERVAL '1 month'
			  FROM months
			  WHERE m < ?::DATE - INTERVAL '1 month'
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
				TO_CHAR(t.date, 'YYYY-MM')   AS ym,
				SUM(t.amount)                  AS amt
			  FROM %s t
			  JOIN %s c ON t.category = c.code
			  LEFT JOIN %s p ON p.name = c.parent
			  WHERE t.date >= ?::DATE AND t.date < ?::DATE
			  GROUP BY gcode, ym
			)
			
			SELECT
			  g.id,
			  g.code AS code,
			  (
				SELECT JSONB_AGG(
					JSONB_BUILD_OBJECT(
						'ym',  TO_CHAR(mon2.m, 'YYYY-MM'),
						'amt', COALESCE(cm2.amt, 0)
					) ORDER BY mon2.m
				)
				FROM months mon2
				LEFT JOIN cat_month cm2
				  ON cm2.gcode = g.code
				 AND cm2.ym = TO_CHAR(mon2.m, 'YYYY-MM')
			  ) AS monthly
			FROM cat_groups g
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
