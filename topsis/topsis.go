package topsis

import "fmt"

func EntropyTopsis(m [][]float64, criterias []bool, options ...int) []int {
	// 0. 原始矩阵
	PrintFloatMatrix("输入矩阵", m)

	// 1. 数据处理
	if printable {
		fmt.Println("1. 数据处理")
	}
	// 1.1 正向化 - 统一指标方向
	ApplyMaximumDirection(m, criterias)
	PrintFloatMatrix("1.1 正向化矩阵：", m)

	// 1.2 归一化 - 去量纲
	ApplyMinMaxScalerNormalization(m)
	PrintFloatMatrix("1.2 归一化矩阵：", m)

	// 2. 计算权重 - 熵权法
	entropyWeights := GetEntropyWeights(m)
	PrintFloatVector("2. 列熵权:", entropyWeights)

	// 3. 加权
	ApplyEntropyWeightsToMatrix(m, entropyWeights)
	PrintFloatMatrix("3. 加权矩阵", m)

	// 4. 确定正负理想解||最优最劣解
	best, worst := GetBestAndWorstIdealVector(m)
	PrintFloatVector("4.1 最优解", best)
	PrintFloatVector("4.2 最劣解", worst)

	// 5. 计算相似度||相对接近度||欧式距离
	dbest, dworst := CulcDistanceToReferencePoints(m, best, worst)
	PrintFloatVector("5.1 最优解距离", dbest)
	PrintFloatVector("5.2 最劣解距离", dworst)

	// 6. 综合评价指数
	scores := CulcScores(dbest, dworst)
	PrintFloatVector("6. 评价指数", scores)

	// 7. 排序
	optionsList := RankScores(scores, options...)
	fmt.Println("排序结果：", optionsList)
	return optionsList
}
