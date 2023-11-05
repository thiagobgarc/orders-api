package handler

import (
	"fmt"
	"net/http"
)

type Order struct{}

func (o *Order) Create(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Create Order")
}

func (o *Order) List(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "List Order")
}

func (o *Order) GetByID(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Get Order by ID")
}

func (o *Order) UpdateByID(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Update Order By ID")
}

func (o *Order) DeleteByID(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Delete Order By ID")
}
