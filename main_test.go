package main_test

import (
	"net/http"
	"testing"

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

func TestUniapi(t *testing.T) {
	service := uniapi.NewService("https://dummyjson.com", uniapi.UnauthenticatedMiddleware{})

	// TODO: Fix to take strings like '/todos/%d' for path?
	service.AddEndpoint(http.MethodGet, &uniapi.BaseEndpoint[TodoList]{
		Path: "/todos",
	})

	todoList, err := uniapi.Call[TodoList](service, http.MethodGet, "/todos", uniapi.Options{})
	if err != nil {
		panic(err)
	}

	t.Logf("%#v", todoList)
}
