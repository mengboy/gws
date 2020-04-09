package gws

import "github.com/google/uuid"

func UUID() string {
	uuid, _ := uuid.NewUUID()
	return uuid.String()
}
