package tinyrbac

import "net/http"

func getHTTPActionOffset(action string) int {
	switch action {
	case http.MethodGet:
		return 0
	case http.MethodPost:
		return 1
	case http.MethodPut:
		return 2
	case http.MethodPatch:
		return 3
	case http.MethodDelete:
		return 4
	default:
		return unknownAction
	}
}
