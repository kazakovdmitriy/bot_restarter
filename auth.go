package main

import "slices"

func IsAllowed(userID int64, allowedIDs []int64) bool {
	return slices.Contains(allowedIDs, userID)
}
