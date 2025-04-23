package models

type Role string

const (
	RoleStudent Role = "student"
	RoleTutor   Role = "tutor"
)

func (r Role) String() string {
	return string(r)
}

func (r Role) IsValid() bool {
	return r == RoleStudent || r == RoleTutor
}
