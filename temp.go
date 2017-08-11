package main

import (
    "fmt"
    "idGenerator/model/cmap"
);

type Application struct {
    idWorkerMap cmap.ConcurrentMap
}

var application Application;

func main() {
    fmt.Println("xxxxxxxx");
    fmt.Println(application);
}
