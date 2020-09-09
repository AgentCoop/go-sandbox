package main
/*
int yyparse();
*/
import "C"
import "math"

//export Pow
func Pow(x float64, y float64) float64 {
	return math.Pow(float64(x), float64(y))
}

func main() {
	C.yyparse()
}