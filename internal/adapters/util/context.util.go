package util

import (
	"context"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"
	"github.com/etwicaksono/data-transaksi-service/internal/constants"
	errorsConst "github.com/etwicaksono/data-transaksi-service/internal/errors"
)

func GetAuthData(ctx context.Context) (authData entity.AuthData, err error) {
	authData, ok := ctx.Value(constants.AuthDataKey).(entity.AuthData)
	if !ok {
		return authData, errorsConst.ErrFailedParseAuthData
	}
	return
}

func SetAuthData(ctx context.Context, authData entity.AuthData) context.Context {
	return context.WithValue(ctx, constants.AuthDataKey, authData)
}
