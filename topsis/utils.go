package topsis

import "fmt"

const printable = false

func NewVector(n int) []float64 {
	return make([]float64, n)
}

func PrintFloatVector(name string, vector []float64) {
	if !printable {
		return
	}
	fmt.Println("===", name, "===")
	for i := 0; i < len(vector); i++ {
		fmt.Printf("%10.3f", vector[i])
	}
	fmt.Println()
}

func PrintFloatMatrix(name string, matrix [][]float64) {
	if !printable {
		return
	}
	fmt.Println("=== ", name, " ===")
	for i := 0; i < len(matrix); i++ {
		for j := 0; j < len(matrix[0]); j++ {
			fmt.Printf("%10.3f", matrix[i][j])
		}

		fmt.Println()
	}
}
