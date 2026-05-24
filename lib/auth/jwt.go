package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/quickbite/analytics-service/lib/appcontext"
)

func VerifyToken(tokenStr, secret string) (*appcontext.Claims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	claims := &appcontext.Claims{
		UserID:      int(mapClaims["userId"].(float64)),
		Role:        mapClaims["role"].(string),
		CountryCode: mapClaims["countryCode"].(string),
	}

	if v, ok := mapClaims["restaurantId"]; ok && v != nil {
		id := int(v.(float64))
		claims.RestaurantID = &id
	}
	if v, ok := mapClaims["restaurantRole"]; ok && v != nil {
		role := v.(string)
		claims.RestaurantRole = &role
	}
	if v, ok := mapClaims["branchIds"]; ok && v != nil {
		if arr, ok := v.([]interface{}); ok {
			ids := make([]int, len(arr))
			for i, item := range arr {
				ids[i] = int(item.(float64))
			}
			claims.BranchIDs = ids
		}
	}

	return claims, nil
}
