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
	protocol := getDBProtocol(a.cfg.Db.Protocol)
	
	transactionTable := model.Transaction{}.TableName()
	categoryTable := model.Category{}.TableName()

	var resumeSQL string
	switch protocol {
	case ProtocolMySQL:
		resumeSQL = fmt.Sprintf(`
			WITH RECURSIVE months(m) AS (
			  SELECT DATE(?)
			  UNION ALL
			  SELECT DATE_ADD(m, INTERVAL 1 MONTH)
			  FROM months
			  WHERE m < DATE_SUB(DATE(?), INTERVAL 1 MONTH)
			),
			cat_month AS (
			  SELECT
				t.category AS category,
				DATE_FORMAT(t.date, '%%Y-%%m') AS ym,
				SUM(t.amount) AS amt
			  FROM %s t
			  WHERE t.date >= DATE(?) AND t.date < DATE(?)
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
			`, transactionTable, categoryTable)

	case ProtocolPostgreSQL:
		resumeSQL = fmt.Sprintf(`
			WITH RECURSIVE months(m) AS (
			  SELECT ?::DATE
			  UNION ALL
			  SELECT m + INTERVAL '1 month'
			  FROM months
			  WHERE m < ?::DATE - INTERVAL '1 month'
			),
			cat_month AS (
			  SELECT
				t.category AS category,
				TO_CHAR(t.date, 'YYYY-MM') AS ym,
				SUM(t.amount) AS amt
			  FROM %s t
			  WHERE t.date >= ?::DATE AND t.date < ?::DATE
			  GROUP BY t.category, ym
			)
			SELECT
				c.id,
				c.code,
				(
					SELECT JSONB_AGG(
						JSONB_BUILD_OBJECT(
							'ym',  TO_CHAR(mon2.m, 'YYYY-MM'),
							'amt', COALESCE(cm2.amt, 0)
						) ORDER BY mon2.m
					)
					FROM months mon2
					LEFT JOIN cat_month cm2
					  ON cm2.category = c.code
					 AND cm2.ym = TO_CHAR(mon2.m, 'YYYY-MM')
				) AS monthly
			FROM %s c
			GROUP BY c.id, c.code
			ORDER BY c.id;
			`, transactionTable, categoryTable)
	}

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
		result[i] = item.ToEntity()
	}

	return
}
