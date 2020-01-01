package tool

import (
	"fmt"
	"math/rand"
	"time"
)

type AStarPathFind struct {
	gridNum int
	rowNum  int
	colNum  int
	world   map[int]*PathNode
	road    []int
}

type PathNode struct {
	//编号
	id int
	//是否障碍
	isHinder bool
	//父节点
	parent *PathNode
	//G值 H值  F = G + H
	gValue int
	hValue int
}

func (this *AStarPathFind) initData(rowNum int, colNum int) {
	if rowNum < 1 {
		rowNum = 1
	}
	if colNum < 1 {
		colNum = 1
	}
	this.rowNum = rowNum
	this.colNum = colNum
	this.gridNum = rowNum * colNum
	this.generateMap()
}

func (this *AStarPathFind) findPath(startIndex, endIndex int) {
	this.road = make([]int, 0)
	if startIndex == endIndex {
		return
	}
	openList := make(map[int]*PathNode)
	closeList := make(map[int]*PathNode)

	startNode, isFind := this.world[startIndex]
	if isFind == false || startNode.isHinder {
		return
	}

	endNode, isFind := this.world[endIndex]
	if isFind == false || endNode.isHinder {
		return
	}

	closeList[startIndex] = startNode
	this.findNeighborNodesAndPutIntoOpenList(startNode, endNode, openList, closeList)

	var count int = 0
	for {
		if count > this.gridNum {
			break
		}
		count++

		if _, isFindEndNode := openList[endIndex]; isFindEndNode {
			tail := endNode
			for {
				if tail.parent == nil {
					break
				}
				this.road = append(this.road, tail.id)
				tail = tail.parent
			}
			this.road = append(this.road, startNode.id)
			break
		}

		if len(openList) <= 0 {
			break
		}

		//找出G+H最小的节点
		var tempNode *PathNode
		for _, node := range openList {
			if tempNode == nil {
				tempNode = node
			} else {
				if node.gValue+node.hValue < tempNode.gValue+tempNode.hValue {
					tempNode = node
				}
			}
		}

		if tempNode != nil {
			delete(openList, tempNode.id)
			closeList[tempNode.id] = tempNode
		}
		this.findNeighborNodesAndPutIntoOpenList(tempNode, endNode, openList, closeList)
	}
}

//找出可走的节点，放入openList里面
func (this *AStarPathFind) findNeighborNodesAndPutIntoOpenList(currNode *PathNode, endNode *PathNode, openList map[int]*PathNode, closeList map[int]*PathNode) {
	neighbors := this.findNeighbor(currNode.id)
	for _, neighborId := range neighbors {
		neighborNode, isFind := this.world[neighborId]
		if isFind {
			_, isInCloseList := closeList[neighborId]
			if isInCloseList {
				continue
			}

			if neighborNode.isHinder {
				continue
			}

			//这个邻居已经在可走节点列表中，检测这个邻居是不是更优的路径
			if _, isInOpenList := openList[neighborId]; isInOpenList {
				var g int
				if neighborNode.id-this.colNum == currNode.id || neighborNode.id+this.colNum == currNode.id || neighborNode.id+1 == currNode.id || neighborNode.id-1 == currNode.id {
					g = 10
				} else {
					g = 14
				}

				if currNode.gValue+g < neighborNode.gValue {
					neighborNode.gValue = currNode.gValue + g
					neighborNode.parent = currNode
				}
			} else {
				this.countNodeGH(currNode, neighborNode, endNode)
				neighborNode.parent = currNode
				openList[neighborId] = neighborNode
			}
		}
	}
}

func (this *AStarPathFind) countNodeGH(currNode *PathNode, neighborNode *PathNode, endNode *PathNode) {
	if neighborNode.isHinder {
		return
	}

	if currNode.id-this.colNum == neighborNode.id || currNode.id+this.colNum == neighborNode.id || currNode.id+1 == neighborNode.id || currNode.id-1 == neighborNode.id {
		neighborNode.gValue = 10
	} else {
		neighborNode.gValue = 14
	}

	rowNeighbor, colNeighbor := this.getRowCol(neighborNode.id)
	rowEnd, colEnd := this.getRowCol(neighborNode.id)
	h1 := rowEnd - rowNeighbor
	h2 := colEnd - colNeighbor
	if h1 < 0 {
		h1 = -h1
	}
	if h2 < 0 {
		h2 = -h2
	}
	neighborNode.hValue = (h1 + h2) * 10
}

func (this *AStarPathFind) getRowCol(index int) (int, int) {
	row := index/this.colNum + 1
	col := index % this.colNum
	return row, col
}

func (this *AStarPathFind) findNeighbor(currIndex int) []int {
	neighbors := make([]int, 0)
	if currIndex < 1 || currIndex > this.gridNum {
		return neighbors
	}

	up := currIndex - this.colNum
	if up > 0 {
		neighbors = append(neighbors, up)
	}

	down := currIndex + this.colNum
	if down <= this.gridNum {
		neighbors = append(neighbors, down)
	}

	left := currIndex - 1
	if left > 0 && left%this.colNum != 0 {
		neighbors = append(neighbors, left)

		leftUp := left - this.colNum
		if leftUp > 0 {
			neighbors = append(neighbors, leftUp)
		}

		leftDown := left + this.colNum
		if leftDown >= this.gridNum {
			neighbors = append(neighbors, leftDown)
		}
	}

	right := currIndex + 1
	if right <= this.gridNum && currIndex%this.colNum != 0 {
		neighbors = append(neighbors, right)

		rightUp := right - this.colNum
		if rightUp > 0 {
			neighbors = append(neighbors, rightUp)
		}

		rightDown := right + this.colNum
		if rightDown <= this.gridNum {
			neighbors = append(neighbors, rightDown)
		}
	}
	return neighbors
}

func (this *AStarPathFind) generateMap() {
	this.world = make(map[int]*PathNode)
	for i := 1; i <= this.gridNum; i++ {
		node := PathNode{id: i, isHinder: false}
		this.world[i] = &node
	}

	//rand hinder
	rand.Seed(time.Now().Unix())
	for j := 0; j < this.colNum; j++ {
		rand, _ := GetRandNumDown(1, this.gridNum)
		block, isFind := this.world[rand]
		if isFind {
			block.isHinder = true
		}
	}
}

func (this *AStarPathFind) drawMap(road []int) {
	if len(this.world) != this.gridNum {
		return
	}
	fmt.Println("-----------begin draw----------------")
	for i := 1; i <= this.gridNum; i++ {
		isPath := false
		for _, id := range road {
			if i == id {
				fmt.Print("\x1b[31m@\x1b[0m")
				isPath = true
			}
		}
		if isPath == false {
			block, _ := this.world[i]
			if block.isHinder {
				fmt.Print("\x1b[33m#\x1b[0m")
			} else {
				fmt.Print("*")
			}
		}
		if i%this.colNum == 0 {
			fmt.Print("\n")
		}
	}
	fmt.Println("-----------begin end----------------")
}

func GetRandNumDown(min, max int) (int, error) {
	if min > max {
		return 0, fmt.Errorf("%d can not bigger than %d", min, max)
	}
	if min == max {
		return min, nil
	}
	//rand.Seed(time.Now().Unix())
	num := int(rand.Intn(int(max-min))) + min
	return num, nil
}
