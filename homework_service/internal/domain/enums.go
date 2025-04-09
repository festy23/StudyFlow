package domain

type UserRole string

const (
	UserRoleStudent UserRole = "student"
	UserRoleTutor   UserRole = "tutor"
	UserRoleAdmin   UserRole = "admin"
)

func (s AssignmentStatus) IsValid() bool {
	switch s {
	case AssignmentStatusUnspecified, AssignmentStatusUnsent,
		AssignmentStatusUnreviewed, AssignmentStatusReviewed, AssignmentStatusOverdue:
		return true
	default:
		return false
	}
}

func ToAssignmentStatus(status string) AssignmentStatus {
	switch status {
	case "UNSENT":
		return AssignmentStatusUnsent
	case "UNREVIEWED":
		return AssignmentStatusUnreviewed
	case "REVIEWED":
		return AssignmentStatusReviewed
	case "OVERDUE":
		return AssignmentStatusOverdue
	default:
		return AssignmentStatusUnspecified
	}
}
