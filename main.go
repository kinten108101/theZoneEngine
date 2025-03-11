package main
import (
	"fmt"
	"log"
	"net/http"
	"thezone/engine/lib/php"
	"runtime"
)
func handleRoot(response http.ResponseWriter, request *http.Request) {
	if runtime.GOOS == "windows" {
		log.Fatalf("This route isn't supported on Windows. Panicking...")
	}
	// Đừng cười t vì sao phải gọi php từ cli, chẳng có thư viện go nào dùng đc
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
