package main
import (
	"fmt"
	"log"
	"net/http"
	"thezone/engine/lib/php"
)
func handleRoot(response http.ResponseWriter, request *http.Request) {
	healthOutput, error := php.Exec("health.php")
	if error != nil {
		log.Fatal(error)
	}
	fmt.Fprintf(response, healthOutput)
}
func main() {
	http.HandleFunc("/", handleRoot)
	log.Fatal(http.ListenAndServe(":8089", nil))
}
