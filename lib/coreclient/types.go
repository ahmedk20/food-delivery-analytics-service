package coreclient

type RolePermission struct {
	Permission string `json:"permission"`
}

type RolePermissionsResponse struct {
	Permissions []RolePermission `json:"permissions"`
}
