package main

import "net/http"

type Route struct {
	Name string
	Method string
	Pattern string
	HandlerFunc http.HandlerFunc
	Restricted bool
}

type Routes []Route

var routes = Routes{
	Route{
		"Index",
		"GET",
		"/",
		Index,
		false,
	},
	Route{
		"AuthHandshake",
		"POST",
		"/auth",
		AuthHandshake,
		false,
	},
	Route{
		"SocketLogin",
		"GET",
		"/login",
		SocketLogin,
		true,
	},
	Route{
		"UpdateUsername",
		"POST",
		"/updatename",
		UpdateUsername,
		true,
	},
	Route{
		"CreateUser",
		"POST",
		"/newuser",
		CreateUserFrontEnd,
		false,
	},
	Route{
		"ListUsers",
		"GET",
		"/users",
		ListUsersFrontEnd,
		false,
	},
}