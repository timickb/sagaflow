package domain

type GetFeedDto struct {
	Count    int64
	Statuses []InstanceStatus
}
