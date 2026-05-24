package appcontext

import "context"

type ctxKey string

const (
	keyCorrelationID ctxKey = "correlation_id"
	keyClaims        ctxKey = "claims"
)

type Claims struct {
	UserID         int      `json:"userId"`
	Role           string   `json:"role"`
	CountryCode    string   `json:"countryCode"`
	RestaurantID   *int     `json:"restaurantId,omitempty"`
	RestaurantRole *string  `json:"restaurantRole,omitempty"`
	BranchIDs      []int    `json:"branchIds,omitempty"`
}

func SetCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyCorrelationID, id)
}

func CorrelationID(ctx context.Context) string {
	if v, ok := ctx.Value(keyCorrelationID).(string); ok {
		return v
	}
	return ""
}

func SetClaims(ctx context.Context, c *Claims) context.Context {
	return context.WithValue(ctx, keyClaims, c)
}

func GetClaims(ctx context.Context) *Claims {
	if v, ok := ctx.Value(keyClaims).(*Claims); ok {
		return v
	}
	return nil
}
