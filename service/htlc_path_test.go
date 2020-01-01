package service

import (
	"fmt"
	"github.com/nickdavies/go-astar/astar"
	"log"
	"math/rand"
	"testing"
	"time"
)

func TestAStar(t *testing.T) {
	rows := 3
	cols := 4

	// Build AStar object from existing
	// PointToPoint configuration
	world := astar.NewAStar(rows, cols)
	//p2p := astar.NewPointToPoint()
	p2p := astar.NewListToPoint(false)
	p2p = astar.NewRowToRow()

	// Make an invincible obsticle at (1,1)
	//world.FillTile(astar.Point{1, 1}, 1)
	world.FillTile(astar.Point{1, 2}, -1)
	world.FillTile(astar.Point{1, 3}, -1)

	// Path from one corner to the other
	source := []astar.Point{astar.Point{0, 0}}
	target := []astar.Point{astar.Point{2, 3}}

	path := world.FindPath(p2p, source, target)

	if path == nil {
		log.Println("no path")
		return
	}
	log.Println(path)
	for path != nil {
		fmt.Printf("At (%d, %d)\n", path.Row, path.Col)
		path = path.Parent
	}
}

func TestAStar2(t *testing.T) {
	road := make([]int32, 0)
	GenerateMap()
	fmt.Println("原地图")
	DrawMap(road)
	fmt.Println("==================")
	road = FindRoad(1, 99)
	if len(road) <= 0 {
		fmt.Println("找不到路")
		return
	}
	fmt.Println("路线图")
	DrawMap(road)
	fmt.Println("==================")
	for index, _ := range road {
		log.Println(GetRowCol(road[len(road)-index-1]))
	}
	fmt.Printf("Road: %v", road)
}

type Block struct {
	//编号
	id int32
	//是否障碍
	isHinder bool
	//父节点
	parent *Block
	//G值 H值  F = G + H
	gValue int32
	hValue int32
}

// 总共格子数
var GRID_NUM int32 = 100

// 行数
var ROW_NUM int32 = 10

// 列数
var COL_NUM int32 = 10

// 地图map
var World map[int32]*Block

// 生成地图 10*10 大小
func GenerateMap() {
	World = make(map[int32]*Block)
	GRID_NUM = ROW_NUM * COL_NUM
	for i := int32(1); i <= GRID_NUM; i++ {
		block := &Block{
			id:       i,
			isHinder: false,
		}
		World[i] = block
	}
	// 随机生成障碍物
	//var tem int32 = 25
	rand.Seed(time.Now().Unix())
	for j := 0; j < 10; j++ {
		rand, _ := GetRandNumDown(1, GRID_NUM)
		block, fd := World[rand]
		if fd {
			block.isHinder = true
		}
		//tem += 10
	}
}

//获取[min, max)之间的随机数
func GetRandNumDown(min, max int32) (int32, error) {
	if min > max {
		return 0, fmt.Errorf("%d can not bigger than %d", min, max)
	}
	if min == max {
		return min, nil
	}
	//rand.Seed(time.Now().Unix())
	num := int32(rand.Intn(int(max-min))) + min
	return num, nil
}

func DrawMap(road []int32) {
	if len(World) != int(GRID_NUM) {
		return
	}
	for i := int32(1); i <= GRID_NUM; i++ {
		isPath := false
		for _, id := range road {
			if i == id {
				fmt.Print("\x1b[31m@\x1b[0m")
				isPath = true
			}
		}
		if isPath == false {
			block, _ := World[i]
			if block.isHinder {
				fmt.Print("\x1b[33m#\x1b[0m")
			} else {
				fmt.Print("*")
			}
		}
		if i%COL_NUM == 0 {
			fmt.Print("\n")
		}
	}
}

//地图寻路 传入起点 终点
func FindRoad(start, end int32) []int32 {
	road := make([]int32, 0)

	if start == end {
		return road
	}
	//开放列表
	OpenList := make(map[int32]*Block)
	//闭合列表
	CloseList := make(map[int32]*Block)

	startNode, isFind := World[start]
	if !isFind || startNode.isHinder {
		return road
	}

	endNode, isFind := World[end]
	if !isFind || endNode.isHinder {
		return road
	}

	// 起点放入closelist
	CloseList[start] = startNode

	Run(startNode, endNode, OpenList, CloseList)
	/////////
	// 在开放列表中找最小值
	var count int32 = 0
	for {
		if count >= 100000 {
			fmt.Println("BREAK")
			break
		}
		count++
		if _, isFind := OpenList[end]; isFind {
			//找到了
			tail := endNode
			for {
				if tail.parent == nil {
					break
				}
				road = append(road, tail.id)
				tail = tail.parent
			}
			break
		}
		//如果开启列表已经为空了，那就是没有出路了，开启列表：可供选的路的节点列表
		if len(OpenList) <= 0 {
			//找不到路
			break
		}

		//找出开启列表中，g+h最小的那个节点，因为花费成本最小，也就是最优路径
		var tempNode *Block
		for _, val := range OpenList {
			if tempNode == nil {
				tempNode = val
			} else {
				if val.gValue+val.hValue < tempNode.gValue+tempNode.hValue {
					tempNode = val
				}
			}
		}
		// 找到后从开放列表中删除, 加入关闭列表
		if tempNode != nil {
			delete(OpenList, tempNode.id)
			CloseList[tempNode.id] = tempNode
		}
		Run(tempNode, endNode, OpenList, CloseList)
	}
	//road = append(road, start)
	return road
}

func Run(currNode, endNode *Block, OpenList, CloseList map[int32]*Block) {
	// 找出邻居
	neighbors := FindNeighbor(currNode.id)
	for _, neighborId := range neighbors {
		neighborBlock, ok := World[neighborId]
		if ok {
			// 已经在关闭列表,不考虑
			_, isFind := CloseList[neighborBlock.id]
			if isFind {
				continue
			}
			// 是障碍 不考虑
			if neighborBlock.isHinder {
				continue
			}
			//如果这个邻居已经在开启列表中，那么就要更新它的g值
			if _, isFind := OpenList[neighborBlock.id]; isFind {
				// 已经在开放列表
				var g int32
				if neighborBlock.id+1 == currNode.id || neighborBlock.id-1 == currNode.id || neighborBlock.id+COL_NUM == currNode.id || neighborBlock.id-COL_NUM == currNode.id {
					g = 10
				} else {
					g = 15
				}
				//因为当前这条路最近，所以更新此邻居的g值，也就是起点到邻居的g值
				if currNode.gValue+g < neighborBlock.gValue {
					neighborBlock.parent = currNode
					neighborBlock.gValue = currNode.gValue + g
				}
			} else { // 如果邻居不在开启列表，计算g和h，放入开启列表，设定其父节点为当前节点，父节点的存在，是为了反推，从终点反推路径到起点
				// 计算每个相邻格子的gh值
				CountGH(currNode, neighborBlock, endNode)
				// 放入开放列表
				OpenList[neighborBlock.id] = neighborBlock
				neighborBlock.parent = currNode
			}
		}
	}
}

// 计算B点 G H 值
func CountGH(pointA, pointB, pointC *Block) {
	if pointB.isHinder {
		return
	}
	rowB, colB := GetRowCol(pointB.id)
	rowC, colC := GetRowCol(pointC.id)
	h1 := rowB - rowC
	h2 := colB - colC
	if h1 < 0 {
		h1 = -h1
	}
	if h2 < 0 {
		h2 = -h2
	}
	pointB.hValue = (h1 + h2) * 10
	if pointA.id-COL_NUM == pointB.id || pointA.id+COL_NUM == pointB.id || pointA.id+1 == pointB.id || pointA.id-1 == pointB.id {
		pointB.gValue = 10
	} else {
		pointB.gValue = 15
	}
}

// 计算给出的位置在哪一行哪一列
func GetRowCol(point int32) (int32, int32) {
	row := point/COL_NUM + 1
	col := point % COL_NUM
	return row, col
}

// 找到相邻格子
func FindNeighbor(point int32) []int32 {
	neighbor := make([]int32, 0)
	if point < 1 || point > GRID_NUM {
		return neighbor
	}
	// 上
	up := point - COL_NUM
	if up > 0 {
		neighbor = append(neighbor, up)
	}
	// 下
	down := point + COL_NUM
	if down <= GRID_NUM {
		neighbor = append(neighbor, down)
	}
	// 左
	left := point - 1
	if left > 0 && left%COL_NUM != 0 {
		neighbor = append(neighbor, left)
		// 左上
		leftup := left - COL_NUM
		if leftup > 0 {
			neighbor = append(neighbor, leftup)
		}
		// 左下
		leftdown := left + COL_NUM
		if leftdown <= GRID_NUM {
			neighbor = append(neighbor, leftdown)
		}
	}
	// 右
	right := point + 1
	if right <= GRID_NUM && point%COL_NUM != 0 {
		neighbor = append(neighbor, right)
		// 右上
		rightup := right - COL_NUM
		if rightup > 0 {
			neighbor = append(neighbor, rightup)
		}
		// 右下
		rightdown := right + COL_NUM
		if rightdown <= GRID_NUM {
			neighbor = append(neighbor, rightdown)
		}
	}
	return neighbor
}
