package forcehalving

import (
	"fmt"
	"testing"
)

func TestHalving(t *testing.T) {
	fmt.Println("testingtesting")
	sol, err := ForceHalve([]int{10, 20})
	fmt.Println(err)
	fmt.Println(sol)
	sol, err = ForceHalve([]int{1, 2, 3, 4, 5, 7, 8, 9, 11, 22, 34})
	fmt.Println(err)
	fmt.Println(sol)
	sol, err = ForceHalve([]int{1, 2, 3, 4})
	fmt.Println(err)
	fmt.Println(sol)
}
