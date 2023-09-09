package system

type SysMenusResponse struct {
	Menus []SysMenu `json:"menus"`
}

type SysBaseMenusResponse struct {
	Menus []SysBaseMenu `json:"menus"`
}

type SysBaseMenuResponse struct {
	Menu SysBaseMenu `json:"menu"`
}
