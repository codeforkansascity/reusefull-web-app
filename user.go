package main

type User struct {
	ID       string
	Name     string
	Email    string
	LoggedIn bool
	Admin    bool
	Charity  int
	CanEdit  bool
}
