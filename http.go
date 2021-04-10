package main

import (
	"fmt"
	"net/http"
	"path"
)

type FileSystemHandler struct {
}

func (f FileSystemHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	fmt.Println(path.Base(request.URL.Path))
	writer.Write([]byte{})
}
