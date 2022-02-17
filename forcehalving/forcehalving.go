package forcehalving

import (
	"fmt"
	"math/big"
)

func ForceHalve(blocks []int) (sol []int, err error) {
	sum := 0
	for _, b := range blocks {
		sum += b
	}
	if sum%2 == 1 {
		return nil, fmt.Errorf("There is no solution")
	}

	half := sum / 2
	two := big.NewInt(2)
	two.Exp(two, big.NewInt(int64(len(blocks))), nil)
	takes := two.Uint64()
	var j uint64
	for j = takes - 1; j > 0; j-- {
		test := 0
		ji := j

		for _, b := range blocks {
			if ji&1 == 1 {
				test += b
			}
			ji >>= 1

		}

		if test == half {
			sol = make([]int, len(blocks))
			for i := 0; i < len(sol); i++ {
				sol[i] = 2*int(j&1) - 1
				j >>= 1
			}

			return
		}
	}

	return nil, fmt.Errorf("There is no solution")

}
