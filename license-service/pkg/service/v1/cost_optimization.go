package v1

import (
	"math"
	"sort"
	"strconv"
)

var (
	acqRightWithHigherValue, acqRightWithLowerValue Pair
	contractSetMap                                  = make(map[string]ContractSet)
)

type Contract struct {
	Memory float64
	Node   float64
	Amount float64
}

type Pair struct {
	Key   int
	Value Contract
}

type PairList []Pair

type ContractSet struct {
	Set    map[int]int
	Amount float64
}

type ContractSetPair struct {
	Key   string
	Value ContractSet
}

type ContractSetPairList []ContractSetPair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PairList) Less(i, j int) bool { return p[i].Value.Amount > p[j].Value.Amount }

func (p ContractSetPairList) Len() int           { return len(p) }
func (p ContractSetPairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p ContractSetPairList) Less(i, j int) bool { return p[i].Value.Amount < p[j].Value.Amount }

// Approch 2, navie approach where it try all solutions selecting the lowest acquiredRight amout with minimum value of highest acquiredRight amount
// Data used for algorithm is 2 acquired right, one with lower price and less number of units(memory, node) and other one is with higher price with more number of units
// Delta value is used as difference acquired through compliance
// 1. As first step calculate the HighestContractByAmount by sorting the higher value price
// 2. Once we get the highest value price, initialize the set[minimumNumberAcquiredRightFromHighest, MaximumNumberOfAcquiredRightFromLowestPrice]
// 3. Start a loop till we reach the 0th value of MaximumNumberOfAcquiredRightFromLowestPrice by increasing minimumNumberAcquiredRightFromHighest by 1.
// 4. once we get all the cases sort the amount with lowest value and return the result
func GetNumberOfAcquiredRightByOptimizingCost(acqRights map[int]Contract, delta [1][2]float64) ContractSetPairList {
	deltaMemory := delta[0][0]
	deltaNode := delta[0][1]

	highestAmountOrder := GetHighestContractByAmount(acqRights)
	// fmt.Println(highestAmountOrder[0].Value.Amount)
	acqRightWithHigherValue = highestAmountOrder[0]
	acqRightWithLowerValue = highestAmountOrder[1]

	// take 0 unit for the acqRightWithHighestOrderValue initialize the first value of lowerAcqRight
	var initailLowerValueRequiredAcqRight, initailHighestValueRequiredAcqRight int
	initailLowerValueRequiredAcqRight = GetMaxmimumRequiredContractValueFromLowestPricedAcquiredRight(acqRightWithLowerValue, deltaMemory, deltaNode)

	for initailLowerValueRequiredAcqRight >= 0 {
		key := "set" + strconv.Itoa(initailLowerValueRequiredAcqRight)
		c := ContractSet{
			Set: map[int]int{},
		}
		c.Set[highestAmountOrder[0].Key] = initailHighestValueRequiredAcqRight
		c.Set[highestAmountOrder[1].Key] = initailLowerValueRequiredAcqRight
		c.Amount = acqRightWithLowerValue.Value.Amount*float64(initailLowerValueRequiredAcqRight) + acqRightWithHigherValue.Value.Amount*float64(initailHighestValueRequiredAcqRight)
		contractSetMap[key] = c

		// Increased initailHighestValueRequiredAcqRight by 1
		initailHighestValueRequiredAcqRight++

		remainingMemoryDelta := deltaMemory - acqRightWithHigherValue.Value.Memory
		remainingNodeDelta := deltaNode - acqRightWithHigherValue.Value.Node

		deltaMemory = remainingMemoryDelta
		deltaNode = remainingNodeDelta

		initailLowerValueRequiredAcqRight = GetMaxmimumRequiredContractValueFromLowestPricedAcquiredRight(acqRightWithLowerValue, remainingMemoryDelta, remainingNodeDelta)
	}

	i := 0
	p := make(ContractSetPairList, len(contractSetMap))

	for k, v := range contractSetMap {
		p[i] = ContractSetPair{Key: k, Value: v}
		i++
	}

	sort.Sort(p)
	return p

}

func GetMaxmimumRequiredContractValueFromLowestPricedAcquiredRight(acqRghtWithLowestPrice Pair, deltaMemory, deltaNode float64) int {
	compliancedMemory := deltaMemory / acqRghtWithLowestPrice.Value.Memory
	complianceNode := deltaNode / acqRghtWithLowestPrice.Value.Node

	if compliancedMemory > complianceNode {
		return int(math.Ceil(compliancedMemory))
	}
	return int(math.Ceil(complianceNode))

}

func GetHighestContractByAmount(acqRights map[int]Contract) PairList {
	i := 0
	p := make(PairList, len(acqRights))

	for k, v := range acqRights {
		p[i] = Pair{Key: k, Value: v}
		i++
	}

	sort.Sort(p)

	// for _, k := range p {
	// 	fmt.Printf("%v\t%v\n", k.Key, k.Value)
	// }
	return p
}
