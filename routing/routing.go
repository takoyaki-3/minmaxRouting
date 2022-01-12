package routing

import (
	// "fmt"

	"fmt"

	"github.com/takoyaki-3/minmaxRouting"
	pb "github.com/takoyaki-3/minmaxRouting/pb"
)

type Query struct {
	NWeight int
	FromNode minmaxrouting.NodeIdType
	ToNode minmaxrouting.NodeIdType
	MaxTransfer int
	IsSerialNG bool
}

type Route struct {
	Nodes []minmaxrouting.NodeIdType
	Weight minmaxrouting.Weight
	SubNodes [][]minmaxrouting.NodeIdType
}

type Memo [][]int

func MinMaxRouting(g *minmaxrouting.Graph,query Query)(routes []Route,memo Memo,memoList []CB){

	toNode := minmaxrouting.NodeIdType(query.ToNode)

	// マーティン拡張による最大・最小の
	memo = make(Memo,len(g.Nodes))

	// 初期化
	initEdge := minmaxrouting.EdgeIdType(-1)
	memoList = []CB{
		CB{},
		CB{
			BeforeEdgeId: initEdge,
			Weight: *minmaxrouting.NewWeight(query.NWeight),
			IsUse: true,
			Node: query.FromNode,
			BeforeCB: -1,
		},
	}
	memo[query.FromNode] = append(memo[query.FromNode], 1)

	cou := 0
	que := Que{}
	que.Add(1)

	for que.Len() > 0 {
		posCBIndex := que.Get()
		posCb := memoList[posCBIndex]

		cou++
		if toNode >= 0 && cou % 1000 == 0 {
			fmt.Println(cou,que.Len(),len(memo[toNode]))
		}

		// posCbが既に未使用のものであればスキップ
		if !posCb.IsUse {
			continue
		}

		// 目的地が存在するクエリの場合、目的地到達判定を行う
		if toNode != -1 {
			// 枝刈り：既知のゴールへの重みより大きいか検証
			if !BetterIndex(posCb.Weight,&memo[toNode],&memoList) {
				posCb.IsUse = false
				continue
			}
			if posCb.Node == toNode {
				memo[toNode] = append(memo[toNode], posCBIndex)
				// fmt.Println("posCb",posCb)

				// for i,p := range memo[toNode]{
				// 	if !BetterIndex(memoList[p].Weight,&memo[toNode],&memoList){
				// 		memo[toNode] = append(memo[toNode][:i], memo[toNode][i+1:]...)
				// 	}
				// }
				continue
			}
		}

		// Posが始点となる辺から探索範囲拡張を行っていく
		for _,edgeId := range g.Nodes[posCb.Node].FromEdgeIds{
			edge := g.Edges[edgeId]

			// 同一辺タイプ（コンポーネント毎に異なるEdgeTypeIdを想定）であればスキップ
			if query.IsSerialNG && posCb.BeforeEdgeId != initEdge && edge.EdgeTypeId == g.Edges[posCb.BeforeEdgeId].EdgeTypeId{
				continue
			}
			
			// 乗換辺を連続使用しない。乗換辺はEdgeTypeIdは-1。mkTreeの時に使用する。
			if posCb.BeforeEdgeId != initEdge && edge.EdgeTypeId == -1 && g.Edges[posCb.BeforeEdgeId].EdgeTypeId == -1{
				continue
			}

			// // 同一路線を排除
			// if !query.IsSerialNG {
			// 	if posCb.BeforeEdgeId != initEdge{
			// 		beforeUseTrips := g.Edges[posCb.BeforeEdgeId].UseTrips
			// 		if len(beforeUseTrips) > 0{
			// 			befS := -1
			// 			for tripIndex,trip := range edge.UseTrips{
			// 				if trip == beforeUseTrips[0]{
			// 					befS = tripIndex
			// 				}
			// 			}
			// 			if befS != -1{
			// 				flag := false
			// 				for index:=0;index<len(edge.UseTrips)-befS;index++{
			// 					if len(beforeUseTrips) == index{
			// 						flag = true
			// 						break
			// 					}
			// 					if edge.UseTrips[befS+index] != beforeUseTrips[index]{
			// 						flag = true
			// 						break
			// 					}
			// 				}
			// 				if !flag {
			// 					continue
			// 				}
			// 			}
			// 		}
			// 	}
			// }

			newW := WeightAdder(posCb.Weight,edge.Weight)

			// 乗換え回数の判定
			if int(newW.Weights[1].Min) > query.MaxTransfer {
				continue
			}

			// 既知のゴールへの重みより大きいか検証
			if toNode != -1 && !BetterIndex(newW,&memo[toNode],&memoList) {
				continue
			}
			f := true
			for index,mi := range memo[edge.ToId]{
				if mi < 0 || !memoList[mi].IsUse{
					continue
				}
				m := memoList[mi]
				cw := CompWeight(m.Weight,newW)
				if cw == -1 {
					f = false
					break
				} else if cw == 1 {
					memoList[mi].IsUse = false
					memo[edge.ToId][index] = -2
				}
			}
			if !f{
				continue
			}

			// 枝刈り施策
			flag := true
			// 辺の経由駅が訪問済みの場合
			for _,n := range edge.ViaNodes{
				if edge.ToId == n{
					flag = false
					break
				}
			}
			if !flag{
				continue
			}

			// 既に訪問済みの頂点であればスキップ
			if posCb.BeforeCB > 0{
				bposcb := memoList[posCb.BeforeCB]
				flag := true
				for {
					if bposcb.BeforeEdgeId == initEdge {
						break
					}
					if bposcb.Node == edge.ToId {
						flag = false
						break
					}
					bposcb = memoList[bposcb.BeforeCB]
				}
				if !flag {
					continue
				}
			}

			cbIndex := len(memoList)
			memoList = append(memoList, CB{
				Node: edge.ToId,
				BeforeEdgeId: edgeId,
				Weight: newW,
				IsUse: true,
				BeforeCB: posCBIndex,
			})

			que.Add(cbIndex)
			memo[edge.ToId] = append(memo[edge.ToId], cbIndex)
		}
	}

	if toNode != -1 {
		for _,m:=range memo[toNode] {
			if m >= 0{
				if BetterIndex(memoList[m].Weight,&memo[toNode],&memoList) {
					r := GetRoute(m,&memoList)
					if r != nil {
						routes = append(routes, *r)
					}
				}
			}
		}
	}

	return routes,memo,memoList
}

func GetRoutes(cbs *[]*CB,memoList *[]CB)[]Route{
	routes := []Route{}
	for _,cb := range *cbs{
		if Better(cb.Weight,cbs) {
			r := GetRoute(cb.BeforeCB,memoList)
			if r != nil{
				routes = append(routes, *r)
			}
		}
	}
	return routes
}

func GetRouteTree(memo Memo,memoList []CB)(*pb.RouteTree){
	tree := &pb.RouteTree{}
	tree.Leaves = make([]*pb.Leaf, len(memoList))
	for index,v:=range memoList{
		wight := &pb.Weight{}
		for _,w := range v.Weight.Weights {
			wight.Max = append(wight.Max, int32(w.Max))
			wight.Min = append(wight.Min, int32(w.Min))
		}
		tree.Leaves[index] = CB2Leaf(&v,wight)
	}
	return tree
}

func CB2Leaf(cb *CB,wight *pb.Weight)*pb.Leaf{
	if cb.BeforeCB == 0{
		cb.BeforeCB = -3
	}
	return &pb.Leaf{
		IsUse: cb.IsUse,
		NodeId: int32(cb.Node),
		Index: int32(cb.BeforeCB),
		BeforeEdgeId: int32(cb.BeforeEdgeId),
		Weight: wight,
	}
}

func GetRoute(pos int,memoList *[]CB)*Route{
	route := Route{
		Weight: (*memoList)[pos].Weight,
	}
	for {
		if !(*memoList)[pos].IsUse {
			return nil
		}
		route.Nodes = append([]minmaxrouting.NodeIdType{(*memoList)[pos].Node},route.Nodes...)
		if (*memoList)[pos].BeforeCB < 0 {
			break
		}
		pos = (*memoList)[pos].BeforeCB
	}
	return &route
}

func Better(weight minmaxrouting.Weight,ws *[]*CB)bool{
	for _,w := range *ws{
		if w != nil{
			if CompWeight(weight,w.Weight) == 1 {
				return false
			}	
		}
	}
	return true
}
func BetterIndex(weight minmaxrouting.Weight,ws *[]int,memoList *[]CB)bool{
	for _,w := range *ws{
		if w >= len(*memoList) {
			fmt.Println("range over",len(*memoList),w)
			continue
		}
		if (*memoList)[w].IsUse {
			if w >= 0 && len((*memoList)[w].Weight.Weights) > 0{
				if CompWeight(weight,(*memoList)[w].Weight) == 1 {
					return false
				}	
			}	
		}
	}
	return true
}

func WeightAdder(w1 minmaxrouting.Weight,w2 minmaxrouting.Weight)(w minmaxrouting.Weight){
	for i,_ := range w1.Weights{
		w.Weights = append(w.Weights, minmaxrouting.MinMax{
			Min: w1.Weights[i].Min+w2.Weights[i].Min,
			Max: w1.Weights[i].Max+w2.Weights[i].Max,
		})
	}
	return w
}

type Que struct {
	Cb []int
}

func (q *Que)Add(cb int){
	q.Cb = append(q.Cb, cb)
}

func (q *Que)Len()int{
	return len(q.Cb)
}

func (q *Que)Get()(int){
	cb := q.Cb[0]
	q.Cb = q.Cb[1:]
	return cb
}

// w1が良い場合-1、w2が良い場合1、一概に決まらない場合0
func CompWeight(w1 minmaxrouting.Weight,w2 minmaxrouting.Weight)int{
	is1Better := true
	is2Better := true

	for i,_:=range w1.Weights{
		if !(w1.Weights[i].Max <= w2.Weights[i].Min) {
			// w1が一概に優っているとは言えない
			is1Better = false
		}
		if !(w2.Weights[i].Max <= w1.Weights[i].Min) {
			// w1が一概に優っているとは言えない
			is2Better = false
		}
	}
	if is1Better == is2Better {
		return 0
	}
	if is1Better {
		return -1
	}
	return 1
}

type CB struct {
	Node minmaxrouting.NodeIdType
	BeforeEdgeId minmaxrouting.EdgeIdType
	Weight minmaxrouting.Weight
	IsUse bool
	BeforeCB int
}

