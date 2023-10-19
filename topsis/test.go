package topsis

import "fmt"

func main() {
	m := [][]float64{
		{1, 2, -3, 4, 5, 6},
		{1, 1, -1, 1, 1, 1},
		{2, 2, -2, 2, 2, 2},
		{4, 4, -4, 4, 4, 4},
	}
	criterias := []bool{true, true, false, true, true}
	//options := []int{1, 2, 3, 4}
	result := EntropyTopsis(m, criterias)
	fmt.Println(result)
}
