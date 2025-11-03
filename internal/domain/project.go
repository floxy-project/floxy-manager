package domain

import (
	"strconv"
	"time"
)

type ProjectID int

type Project struct {
	ID          ProjectID
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ArchivedAt  *time.Time
}

type ProjectDTO struct {
	Name        string
	Description string
}

func (id ProjectID) String() string {
	return strconv.Itoa(int(id))
}

func (id ProjectID) Int() int {
	return int(id)
}
