package domain

// InstancesFeed - результат запроса страницы фида инстансов
type InstancesFeed struct {
	Instances       []*InstanceView
	PaginationToken string
}
