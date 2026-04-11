package handlers

import (
	"database/sql"
	"time"
	"github.com/google/uuid"
)

func toNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func toNullInt32(i *int32) sql.NullInt32 {
	if i == nil {
		return sql.NullInt32{Valid: false}
	}
	return sql.NullInt32{Int32: *i, Valid: true}
}

func toNullUUID(u string) uuid.NullUUID {
	id, err := uuid.Parse(u)
	if err != nil {
		return uuid.NullUUID{Valid: false}
	}
	return uuid.NullUUID{UUID: id, Valid: true}
}

func parseDate(s string) (sql.NullTime, error) {
	if s == "" {
		return sql.NullTime{Valid: false}, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		t, err = time.Parse(time.RFC3339, s)
		if err != nil {
			return sql.NullTime{Valid: false}, err
		}
	}
	return sql.NullTime{Time: t, Valid: true}, nil
}


