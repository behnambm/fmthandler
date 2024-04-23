# fmthandler

fmthandler is a tool that enhances Go HTTP server code by automatically adding logging to HTTP HandlerFunc handlers. 
It parses the Abstract Syntax Tree (AST) of your code to achieve this.


### Features
- Automatically adds logging statements to HTTP HandlerFunc handlers.
- Works seamlessly with both single files and entire directories of Go code.


## Installation 
```shell
go install github.com/behnambm/fmthandler@latest
```


### Code Before
```go 
package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("pong"))
	})

	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatalln(err)
	}

}

```

### Running the fmthanlder 
```shell
 fmthandler --file main.go
```

### Code After
```go 
package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Invoking HandlerFunc: '/ping'")
		w.WriteHeader(200)
		w.Write([]byte("pong"))
	})

	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatalln(err)
	}

}

```
