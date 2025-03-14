package main

import (
	"log"
	"net/http"

	"github.com/Alexamakans/uniapi/pkg/uniapi"
)

type Todo struct {
	Id        int    `json:"id"`
	Todo      string `json:"todo"`
	Completed bool   `json:"completed"`
	UserId    int    `json:"userId"`
}

type TodoList struct {
	Todos []Todo `json:"todos"`
	Total int    `json:"total"`
	Skip  int    `json:"skip"`
	Limit int    `json:"limit"`
}

func main() {
	service := uniapi.NewService("https://dummyjson.com", uniapi.UnauthenticatedMiddleware{})

	// TODO: Fix to take strings like '/todos/%d' for path?
	service.AddEndpoint(http.MethodGet, &uniapi.BaseEndpoint[TodoList]{
		Path: "/todos",
		Paginator: &uniapi.SkipLimitPaginator{
			CountFieldName: "Total",
			SkipFieldName:  "Skip",
			LimitFieldName: "Limit",
			ListFieldName:  "Todos",

			SkipQueryName:  "skip",
			LimitQueryName: "limit",
		},
	})

	todoList, err := uniapi.Call[TodoList](service, http.MethodGet, "/todos", uniapi.Options{})
	if err != nil {
		panic(err)
	}

	log.Printf("%#v", todoList)
}
