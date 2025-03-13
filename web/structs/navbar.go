package structs

type NavBar struct {
	Titulo string
	Itens  []NavItem
}

type NavItem struct {
	Texto string
	Link  string
}
