package topsis

import (
	"fmt"
	"math"
)

// MaximumDirection 正向化
func ApplyMaximumDirection(m [][]float64, d []bool) {
	for i := range d {
		if d[i] {
			continue
		}
		for j := range m {
			m[j][i] = -m[j][i]
		}
	}
}

// ApplyMinMaxScalerNormalization x' = (x - min) / (max - min)
func ApplyMinMaxScalerNormalization(m [][]float64) [][]float64 {
	rowsNum := len(m)
	colsNum := len(m[0])
	// find min max from each column
	for j := 0; j < colsNum; j++ {
		min := m[0][j]
		max := m[0][j]
		// find min max in column
		for i := 0; i < rowsNum; i++ {
			t := m[i][j]
			if t < min {
				min = t
			}
			if t > max {
				max = t
			}
		}
		// min-max-scale
		for i := 0; i < rowsNum; i++ {
			m[i][j] = (m[i][j] - min) / (max - min)
		}
	}
	return m
}

func GetEntropyWeights(m [][]float64) []float64 {
	rowsNum := len(m)
	colsNum := len(m[0])

	entropies := make([]float64, colsNum)
	for j := 0; j < colsNum; j++ {
		sum := 0.0
		for i := 0; i < rowsNum; i++ {
			sum += m[i][j]
		}

		entropy := 0.0
		for i := 0; i < rowsNum; i++ {
			p := m[i][j] / sum
			if p > 0 {
				entropy += -p * math.Log2(p)
			}
		}

		entropies[j] = entropy
	}
	return entropies
}

func ApplyEntropyWeightsToMatrix(m [][]float64, weights []float64) [][]float64 {
	for j, w := range weights {
		for i := range m {
			m[i][j] *= w
		}
	}
	return m
}

func GetBestAndWorstIdealVector(m [][]float64) ([]float64, []float64) {
	rowsNum := len(m)
	colsNum := len(m[0])
	bestIdeal := NewVector(colsNum)
	worstIdeal := NewVector(colsNum)

	// for by column
	for j := 0; j < colsNum; j++ {
		max := m[0][j]
		min := m[0][j]
		// find min max in column
		for i := 0; i < rowsNum; i++ {
			t := m[i][j]
			if t < min {
				min = t
			}
			if t > max {
				max = t
			}
		}

		bestIdeal[j] = max
		worstIdeal[j] = min
	}

	return bestIdeal, worstIdeal
}

// CulcDistanceToReferencePoints
//
//	Si+ = √(∑(xij - pij)^2)
//	Si- = √(∑(xij - nj)^2)
func CulcDistanceToReferencePoints(m [][]float64, best []float64, worst []float64) ([]float64, []float64) {
	dbest := make([]float64, len(m))
	dworst := make([]float64, len(m))

	for i := range m {
		simBest := 0.0
		simWorst := 0.0
		for j := range m[i] {
			simBest += math.Pow(m[i][j]-best[j], 2)
			simWorst += math.Pow(m[i][j]-worst[j], 2)
		}
		dbest[i] = math.Sqrt(simBest)
		dworst[i] = math.Sqrt(simWorst)
	}

	return dbest, dworst
}

// CulcScores  Ci = Si- / (Si+ + Si-)
func CulcScores(dbest []float64, dworst []float64) []float64 {
	scores := make([]float64, len(dbest))

	for i := range dbest {
		scores[i] = dworst[i] / (dbest[i] + dworst[i])
	}
	return scores
}

// 分区
func partition(arr []float64, index []int, left, right int) int {
	i := left - 1
	pivot := arr[right]
	for j := left; j < right; j++ {
		if arr[j] >= pivot {
			i++
			arr[i], arr[j] = arr[j], arr[i]
			index[i], index[j] = index[j], index[i]
		}
	}
	i++
	arr[i], arr[right] = arr[right], arr[i]
	index[i], index[right] = index[right], index[i]
	return i
}

// 快速排序
func quickSort(arr []float64, index []int, left, right int) {
	if left < right {
		pivot := partition(arr, index, left, right)
		quickSort(arr, index, left, pivot-1)
		quickSort(arr, index, pivot+1, right)
	}
}

// RankScores
// input: score[] , optionsIndex[]
// output: []int - 评分后的索引清单
func RankScores(s []float64, options ...int) []int {
	indexes := make([]int, len(s)) // 返回评分索引  是对metrics的索引，而不是实际peerId
	arr := make([]float64, len(s))
	for i := range s { // 生成索引列表
		indexes[i] = i
		arr[i] = s[i]
	}
	// 快排 - 对生成的索引进行排序
	quickSort(arr, indexes, 0, len(arr)-1)

	// 如果有额外的实际peerId  那么就将indexes映射为peers
	if len(options) == 0 {
		return indexes
	} else {
		newOptions := make([]int, len(options)) // 新的序列
		fmt.Println(len(indexes), indexes)
		fmt.Println(len(options), options)
		for i := range indexes {
			newOptions[i] = options[indexes[i]]
		}
		return newOptions // index == order, value == peerId
	}
}
